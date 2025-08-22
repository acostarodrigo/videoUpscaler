package videoUpscaler

import (
	"context"
	fmt "fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	videoUpscalerCrypto "github.com/janction/videoUpscaler/crypto"
	"github.com/janction/videoUpscaler/db"
	"github.com/janction/videoUpscaler/ipfs"
	"github.com/janction/videoUpscaler/videoUpscalerLogger"
	"github.com/janction/videoUpscaler/vm"
)

func (t *VideoUpscalerThread) StartWork(ctx context.Context, worker string, cid string, path string, db *db.DB) error {
	// ctx := context.Background()

	if err := db.UpdateThread(t.ThreadId, false, false, true, false, false, false, false, false); err != nil {
		videoUpscalerLogger.Logger.Error("Unable to update thread status, err: %s", err.Error())
	}

	isRunning := vm.IsContainerRunning(ctx, t.ThreadId)
	if !isRunning {
		// task is not running,

		// we remove the container just in case it already exists.
		vm.RemoveContainer(ctx, "myBlender"+t.ThreadId)

		videoUpscalerLogger.Logger.Info("No solution for thread %s. Starting work", t.ThreadId)
		// we don't have a solution, start working
		started := time.Now().Unix()
		ipfs.EnsureIPFSRunning()
		db.AddLogEntry(t.ThreadId, fmt.Sprintf("Started downloading IPFS file %s...", cid), started, 0)
		if err := db.UpdateThread(t.ThreadId, true, false, true, false, false, false, false, false); err != nil {
			videoUpscalerLogger.Logger.Error("Unable to update thread status, err: %s", err.Error())
		}
		err := ipfs.IPFSGet(cid, path)
		if err != nil {
			if err := db.UpdateThread(t.ThreadId, true, false, true, false, false, false, false, false); err != nil {
				videoUpscalerLogger.Logger.Error("Unable to update thread status, err: %s", err.Error())
			}
			db.AddLogEntry(t.ThreadId, fmt.Sprintf("Error getting IPFS file %s. %s", cid, err.Error()), started, 2)
			videoUpscalerLogger.Logger.Error("Error getting cid %s", cid)
			return err
		}
		// download completed successfuly
		if err := db.UpdateThread(t.ThreadId, true, true, true, false, false, false, false, false); err != nil {
			videoUpscalerLogger.Logger.Error("Unable to update thread status, err: %s", err.Error())
		}

		finish := time.Now().Unix()
		difference := time.Unix(finish, 0).Sub(time.Unix(started, 0))
		db.AddLogEntry(t.ThreadId, fmt.Sprintf("Successfully downloaded IPFS file %s in %v seconds.", cid, int(difference.Seconds())), finish, 0)

		// we start upscaler
		vm.RenderVideo(ctx, cid, t.StartFrame, t.EndFrame, t.ThreadId, path, t.IsReverse(worker), db)

		rendersPath := filepath.Join(path, "output")
		_, err = os.Stat(rendersPath)
		finish = time.Now().Unix()
		difference = time.Unix(finish, 0).Sub(time.Unix(started, 0))
		if err != nil {
			// output path was not created so no upscaler happened. we will start over
			videoUpscalerLogger.Logger.Error("Unable to complete upscaler of task, retrying. No files at %s", rendersPath)

			db.UpdateThread(t.ThreadId, true, true, false, false, false, false, false, false)
			return nil
		}
		files, _ := os.ReadDir(rendersPath)
		if len(files) != int(t.EndFrame)-int(t.StartFrame)+1 {
			db.UpdateThread(t.ThreadId, true, true, false, false, false, false, false, false)
			videoUpscalerLogger.Logger.Error("Not the amount we expected. retrying. Amount of files %v", len(files))
			return nil
		}
		db.UpdateThread(t.ThreadId, true, true, true, true, false, false, false, false)
		db.AddLogEntry(t.ThreadId, fmt.Sprintf("Thread %s completed succesfully in %v seconds.", t.ThreadId, int(difference.Seconds())), finish, 1)
	} else {
		// Container is running, so we update worker status
		// if worker status is idle, we change it
		videoUpscalerLogger.Logger.Info("Work for thread %s is already going", t.ThreadId)

	}

	return nil
}

func (t VideoUpscalerThread) ProposeSolution(codec codec.Codec, alias, workerAddress string, rootPath string, db *db.DB) error {
	db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)

	output := path.Join(rootPath, "renders", t.ThreadId, "output")
	count := vm.CountFilesInDirectory(output)

	if count != (int(t.EndFrame)-int(t.StartFrame))+1 {
		videoUpscalerLogger.Logger.Error("not enought local frames to propose solution: %v", count)
		db.UpdateThread(t.ThreadId, true, true, true, true, false, false, false, false)
		return nil
	}

	hashes, err := GenerateDirectoryFileHashes(output)
	if err != nil {
		videoUpscalerLogger.Logger.Error("Unable to calculate CIDs: %s", err.Error())
		db.UpdateThread(t.ThreadId, true, true, true, true, false, false, false, false)
		return err
	}

	pkey, err := videoUpscalerCrypto.ExtractPublicKey(rootPath, alias, codec)
	if err != nil {
		videoUpscalerLogger.Logger.Error("Unable to extract public key for alias %s at path %s: %s", alias, rootPath, err.Error())
		db.UpdateThread(t.ThreadId, true, true, true, true, false, false, false, false)
		return err
	}

	publicKey := videoUpscalerCrypto.EncodePublicKeyForCLI(pkey)

	for filename, hash := range hashes {
		sigMsg, err := videoUpscalerCrypto.GenerateSignableMessage(hash, workerAddress)

		if err != nil {
			videoUpscalerLogger.Logger.Error("Unable to generate message for worker %s and hash %s: %s", workerAddress, hash, err.Error())
			db.UpdateThread(t.ThreadId, true, true, true, true, false, false, false, false)
			return err
		}

		signature, _, err := videoUpscalerCrypto.SignMessage(rootPath, alias, sigMsg, codec)

		if err != nil {
			videoUpscalerLogger.Logger.Error("Unable to sign message for worker %s and hash %s: %s", workerAddress, hash, err.Error())
			db.UpdateThread(t.ThreadId, true, true, true, true, false, false, false, false)
			return err
		}
		// We rewrite the hash with the signature
		hashes[filename] = videoUpscalerCrypto.EncodeSignatureForCLI(signature)
	}

	solution := MapToKeyValueFormat(hashes)

	// Base arguments
	args := []string{
		"tx", "videoUpscaler", "propose-solution",
		t.TaskId, t.ThreadId,
	}

	// Append solution arguments
	args = append(args, publicKey)
	args = append(args, solution...)

	// Append flags
	args = append(args, "--yes", "--from", workerAddress)
	err = ExecuteCli(args)
	if err != nil {
		videoUpscalerLogger.Logger.Error(err.Error())
		return err
	}

	db.AddLogEntry(t.ThreadId, "Solution proposed. Wainting confirmation...", time.Now().Unix(), 0)

	return nil
}

func (t VideoUpscalerThread) SubmitVerification(codec codec.Codec, alias, workerAddress string, rootPath string, db *db.DB) error {
	// we will verify any file we already have rendered.
	db.UpdateThread(t.ThreadId, true, true, true, true, true, true, false, false)
	output := path.Join(rootPath, "renders", t.ThreadId, "output")
	files := vm.CountFilesInDirectory(output)
	if files == 0 {
		videoUpscalerLogger.Logger.Error("found %v files in path %s", files, output)
		db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)
		return nil
	}

	// Before we calculate verification, we need to make sure we have enought rendered files to submit one.
	totalFiles := t.EndFrame - t.StartFrame
	threshold := float64(totalFiles) * 0.2 // TODO this percentage should be in params

	if float64(files) > threshold {
		videoUpscalerLogger.Logger.Info("rendered files %v at %sis enought to generate verification", files, output)
	} else {
		videoUpscalerLogger.Logger.Error("not enought files %v at %s to generate validation. Upscaler should continue", files, output)
		db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)
		return nil
	}
	// we do have some work, lets compare it with the solution
	myWork, err := GenerateDirectoryFileHashes(output)

	if err != nil {
		videoUpscalerLogger.Logger.Error("error getting hashes. Err: %s", err.Error())
		db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)
		return err
	}

	publicKey, err := videoUpscalerCrypto.GetPublicKey(rootPath, alias, codec)
	if err != nil {
		videoUpscalerLogger.Logger.Error("Error getting public key for alias %s at path %s: %s", alias, rootPath, err.Error())
		db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)
		return err
	}

	for filename, hash := range myWork {
		message, err := videoUpscalerCrypto.GenerateSignableMessage(hash, workerAddress)
		if err != nil {
			videoUpscalerLogger.Logger.Error("unable to generate message to sign %s: %s", message, err.Error())
			db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)
			return err
		}

		signature, _, err := videoUpscalerCrypto.SignMessage(rootPath, alias, message, codec)

		if err != nil {
			videoUpscalerLogger.Logger.Error("unable to sign message %s: %s", message, err.Error())
			db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)
			return err
		}
		// we replace the hash for the signature
		myWork[filename] = videoUpscalerCrypto.EncodeSignatureForCLI(signature)
	}

	db.AddLogEntry(t.ThreadId, "Starting verification of solution...", time.Now().Unix(), 0)

	err = submitValidation(workerAddress, t.TaskId, t.ThreadId, videoUpscalerCrypto.EncodePublicKeyForCLI(publicKey), MapToKeyValueFormat(myWork))

	if err != nil {
		videoUpscalerLogger.Logger.Error("error sending verification: %s", err.Error())
		db.UpdateThread(t.ThreadId, true, true, true, true, true, false, false, false)
		return err
	}
	db.AddLogEntry(t.ThreadId, "Solution verified", time.Now().Unix(), 0)
	return nil
}

func submitValidation(validator string, taskId, threadId, publicKey string, signatures []string) error {
	// Base arguments
	args := []string{
		"tx", "videoUpscaler", "submit-validation",
		taskId, threadId,
	}
	args = append(args, publicKey)
	args = append(args, signatures...)
	args = append(args, "--from")
	args = append(args, validator)
	args = append(args, "--yes")
	err := ExecuteCli(args)
	if err != nil {
		return err
	}
	return nil
}

func (t VideoUpscalerThread) SubmitSolution(ctx context.Context, workerAddress, rootPath string, db *db.DB) error {
	db.UpdateThread(t.ThreadId, true, true, true, true, true, true, true, true)

	db.AddLogEntry(t.ThreadId, "Submiting solution to IPFS...", time.Now().Unix(), 0)
	cid, err := ipfs.UploadSolution(ctx, rootPath, t.ThreadId)
	if err != nil {
		db.UpdateThread(t.ThreadId, true, true, true, true, true, true, true, false)
		videoUpscalerLogger.Logger.Error(err.Error())
		return err
	}

	// we get the average duration of upscaler the frames
	duration, err := db.GetAverageRenderTime(t.ThreadId)
	if err != nil {
		duration = 0
	}

	err = submitSolution(workerAddress, t.TaskId, t.ThreadId, cid, int64(duration))
	if err != nil {
		db.UpdateThread(t.ThreadId, true, true, true, true, true, true, true, false)
		db.AddLogEntry(t.ThreadId, fmt.Sprintf("Error submitting solution. %s", err.Error()), time.Now().Unix(), 2)
		return err
	}

	db.AddLogEntry(t.ThreadId, "Solution uploaded to IPFS correctly.", time.Now().Unix(), 0)
	return nil
}

func submitSolution(address, taskId, threadId string, cid string, duration int64) error {
	args := []string{
		"tx", "videoUpscaler", "submit-solution",
		taskId, threadId,
	}

	// Append solution arguments
	args = append(args, cid)
	args = append(args, strconv.FormatInt(duration, 10))

	// Append flags
	args = append(args, "--yes", "--from", address)

	err := ExecuteCli(args)
	if err != nil {
		return err
	}
	return nil
}

func (t VideoUpscalerThread) IsReverse(worker string) bool {
	for i, v := range t.Workers {
		if v == worker {
			return i%2 != 0
		}
	}
	return false
}

func (t *VideoUpscalerThread) GetValidatorReward(worker string, totalReward types.Coin) types.Coin {
	var totalFiles int
	for _, validation := range t.Validations {
		totalFiles = totalFiles + int(len(validation.Frames))
	}
	for _, validation := range t.Validations {
		if validation.Validator == worker {
			amount := calculateValidatorPayment(int(len(validation.Frames)), totalFiles, totalReward.Amount)
			return types.NewCoin("jct", amount)
		}
	}
	return types.NewCoin("jct", math.NewInt(0))
}

// Calculate the validator's reward proportionally using sdkmath.Int
func calculateValidatorPayment(filesValidated, totalFilesValidated int, totalValidatorReward math.Int) math.Int {
	if totalFilesValidated == 0 {
		return math.NewInt(0) // Avoid division by zero
	}

	// (filesValidated * totalValidatorReward) / totalFilesValidated
	return totalValidatorReward.Mul(math.NewInt(int64(filesValidated))).Quo(math.NewInt(int64(totalFilesValidated)))
}

// Once validations are ready, we show blockchain the solution
func (t *VideoUpscalerThread) RevealSolution(rootPath string, db *db.DB) error {
	output := path.Join(rootPath, "renders", t.ThreadId, "output")
	cids, err := ipfs.CalculateCIDs(output)
	if err != nil {
		videoUpscalerLogger.Logger.Error(err.Error())
		return err
	}

	solution := make(map[string]VideoUpscalerThread_Frame)
	for filename, cid := range cids {
		path := filepath.Join(output, filename)
		hash, err := CalculateFileHash(path)

		if err != nil {
			videoUpscalerLogger.Logger.Error(err.Error())
			return err
		}

		frame := VideoUpscalerThread_Frame{Filename: filename, Cid: cid, Hash: hash}
		solution[filename] = frame
	}

	// Base arguments
	args := []string{
		"tx", "videoUpscaler", "reveal-solution",
		t.TaskId, t.ThreadId,
	}
	args = append(args, FromFramesToCli(solution)...)
	args = append(args, "--from")
	args = append(args, t.Solution.ProposedBy)
	args = append(args, "--yes")
	err = ExecuteCli(args)

	if err != nil {
		return err
	}
	err = db.UpdateThread(t.ThreadId, true, true, true, true, true, true, true, false)
	if err != nil {
		return err
	}
	return nil
}

// Evaluates if the verifications sent are valid
func (t *VideoUpscalerThread) EvaluateVerifications() error {
	for _, frame := range t.Solution.Frames {
		for _, validation := range t.Validations {
			idx := slices.IndexFunc(validation.Frames, func(f *VideoUpscalerThread_Frame) bool { return f.Filename == frame.Filename })

			if idx < 0 {
				// This verification doesn't have the frame of the solution, we skip it hoping another validation has it
				videoUpscalerLogger.Logger.Debug("Solution Frame %s, not found at validation of validator %s ", frame.Filename, validation.Validator)
				continue
			}

			pk, err := videoUpscalerCrypto.DecodePublicKeyFromCLI(validation.PublicKey)
			if err != nil {
				videoUpscalerLogger.Logger.Error("unable to get public key from cli: %s", err.Error())
				return err
			}

			message, err := videoUpscalerCrypto.GenerateSignableMessage(frame.Hash, validation.Validator)
			if err != nil {
				videoUpscalerLogger.Logger.Error("unable to recreate original message %sto verify: %s", message, err.Error())
				return err
			}
			sig, err := videoUpscalerCrypto.DecodeSignatureFromCLI(validation.Frames[idx].Signature)
			if err != nil {
				videoUpscalerLogger.Logger.Error("unable to decode signature: %s", err.Error())
				return err
			}

			valid := pk.VerifySignature(message, sig)

			if valid {
				// verification passed
				frame.ValidCount++
			} else {
				videoUpscalerLogger.Logger.Debug("Verification for frame %s from pk %s NOT VALID!\nMessage: Hash: %s, address: %s\npublicKey:%s\nsignature:%s", validation.Frames[idx].Filename, validation.Validator, frame.Hash, validation.Validator, validation.PublicKey, validation.Frames[idx].Signature)
				frame.InvalidCount++
			}
		}

	}
	return nil
}

// for those frames evaluated, if we have at least one that has more
// invalid counts than valid ones, we rejected. Otherwise is accepted
func (t *VideoUpscalerThread) IsSolutionAccepted() bool {
	validFrameCount := 0

	minValidValidations := 2
	if len(t.Workers) == 1 {
		minValidValidations = 1
	}

	totalFrames := len(t.Solution.Frames)
	if totalFrames == 0 {
		return false // no frames to evaluate
	}

	for _, frame := range t.Solution.Frames {
		if int(frame.ValidCount) >= minValidValidations {
			validFrameCount++
		}
	}

	// Require at least 20% of frames to be valid
	required := int(float64(totalFrames) * 0.2)
	if required == 0 && totalFrames > 0 {
		required = 1 // always require at least 1 if there are frames
	}

	return validFrameCount >= required
}

// validates the IPFS dir contains all files in the solution
func (t *VideoUpscalerThread) VerifySubmittedSolution(dir string) error {
	files, err := ipfs.ListDirectory(dir)
	if err != nil {
		videoUpscalerLogger.Logger.Error("VerifySubmittedSolution dir: %s:%s", dir, err.Error())
		return err
	}
	for _, frame := range t.Solution.Frames {
		if frame.Cid != files[frame.Filename] {
			err := fmt.Errorf("frame %s [%s] doesn't exists in %s", frame.Filename, frame.Cid, dir)
			videoUpscalerLogger.Logger.Error(err.Error())
			return err
		} else {
			videoUpscalerLogger.Logger.Debug("VerifySubmittedSolution file: %s [%s] exists in dir %s", frame.Filename, frame.Cid, dir)
		}

	}
	return nil
}
