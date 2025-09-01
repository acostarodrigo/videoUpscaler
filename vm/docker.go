package vm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/janction/videoUpscaler/db"
	"github.com/janction/videoUpscaler/videoUpscalerLogger"
)

func IsContainerRunning(ctx context.Context, threadId string) bool {
	name := fmt.Sprintf("upscaler-cpu%s", threadId)

	// Command to check for running containers
	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Names}}")

	output, err := cmd.Output()
	if err != nil {
		videoUpscalerLogger.Logger.Error("Error executing Docker command: %v\n", err)
		return false
	}

	// Trim output and compare with container name
	containerName := strings.TrimSpace(string(output))
	return containerName == name
}

func RenderVideo(ctx context.Context, cid string, start int64, end int64, id string, path string, reverse bool, db *db.DB) {
	if reverse {
		for i := end; i >= start; i-- {
			videoUpscalerLogger.Logger.Info("Upscaler` frame %v in reverse", i)
			renderVideoFrame(ctx, cid, i, id, path, db)
		}
	} else {
		for i := start; i <= end; i++ {
			videoUpscalerLogger.Logger.Info("Upscaler frame %v", i)
			renderVideoFrame(ctx, cid, i, id, path, db)
		}
	}
}

func renderVideoFrame(ctx context.Context, cid string, frameNumber int64, id string, path string, db *db.DB) error {
	n := "upscaler-cpu" + id

	started := time.Now().Unix()
	db.AddLogEntry(id, fmt.Sprintf("Started upscaler frame %v...", frameNumber), started, 0)

	// Check if the container exists using `docker ps -a`
	checkCmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", n), "--format", "{{.Names}}")
	output, err := checkCmd.Output()
	if err != nil {
		db.AddLogEntry(id, "Error trying to verify if container already exists.", started, 2)
		fail := fmt.Errorf("failed to check container existence: %w", err)
		videoUpscalerLogger.Logger.Error(fail.Error())
		return fail
	}

	// If the container already exists, exit the function
	if string(output) != "" {
		videoUpscalerLogger.Logger.Debug("Container already exists.")
		return nil
	}

	// Construct the bind path and command
	bindPath := fmt.Sprintf("%s:/work", path)

	var blenderArgs []string

	blenderArgs = append(blenderArgs, "-i")
	blenderArgs = append(blenderArgs, fmt.Sprintf("/work/%s", cid))
	blenderArgs = append(blenderArgs, "-o")
	blenderArgs = append(blenderArgs, "/work/output")
	blenderArgs = append(blenderArgs, "--frame")
	blenderArgs = append(blenderArgs, strconv.FormatInt(frameNumber, 10))
	blenderArgs = append(blenderArgs, "-s")
	blenderArgs = append(blenderArgs, "2")
	blenderArgs = append(blenderArgs, "-n")
	blenderArgs = append(blenderArgs, "-1")
	blenderArgs = append(blenderArgs, "--fast")

	var dockerArgs []string
	dockerArgs = append(dockerArgs, "run")

	dockerArgs = append(dockerArgs, "--rm")
	dockerArgs = append(dockerArgs, "--name")
	dockerArgs = append(dockerArgs, n)
	dockerArgs = append(dockerArgs, "-v")
	dockerArgs = append(dockerArgs, bindPath)
	dockerArgs = append(dockerArgs, "-d")

	// TODO if on Mac, we use another image that is non deterministic
	dockerArgs = append(dockerArgs, "rodrigoa77/upscaler-cpu")

	dockerArgs = append(dockerArgs, blenderArgs...)

	// Create and start the container
	runCmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	videoUpscalerLogger.Logger.Info("Starting docker: %s", runCmd.String())
	err = runCmd.Run()
	if err != nil {
		db.AddLogEntry(id, fmt.Sprintf("Error in creating the container. %s", err.Error()), started, 1)
		videoUpscalerLogger.Logger.Error("failed to create and start container: %s", err.Error())
		return fmt.Errorf("failed to create and start container: %w", err)
	}

	// Wait for the container to finish
	waitCmd := exec.CommandContext(ctx, "docker", "wait", n)
	err = waitCmd.Run()
	if err != nil {
		videoUpscalerLogger.Logger.Error("failed to wait for container: %s", err.Error())
		return fmt.Errorf("failed to wait for container: %w", err)
	}

	// Retrieve and print logs
	logsCmd := exec.CommandContext(ctx, "docker", "logs", n)
	logsOutput, err := logsCmd.Output()
	if err != nil {
		videoUpscalerLogger.Logger.Error("failed to retrieve container logs: %s", err.Error())
		return fmt.Errorf("failed to retrieve container logs: %w", err)
	}
	videoUpscalerLogger.Logger.Info("Container logs:")
	videoUpscalerLogger.Logger.Info(string(logsOutput))

	RemoveContainer(ctx, n)

	// Verify the frame exists and log
	frameFile := FormatFrameFilename(int(frameNumber))
	framePath := filepath.Join(path, "output", frameFile)
	finish := time.Now().Unix()
	difference := time.Unix(finish, 0).Sub(time.Unix(started, 0))
	if _, err := os.Stat(framePath); errors.Is(err, os.ErrNotExist) {
		db.AddLogEntry(id, fmt.Sprintf("Error while upscaler frame %v. %s file is not there", frameNumber, framePath), started, 2)
		renderVideoFrame(ctx, cid, frameNumber, id, path, db)
	} else {
		// we capture the duration of the upscaler
		duration := int(difference.Seconds())
		// we add the log
		db.AddLogEntry(id, fmt.Sprintf("Successfully rendered frame %v in %v seconds.", frameNumber, duration), finish, 1)
		// and record the duration for the frame
		videoUpscalerLogger.Logger.Info("Recorded duration for frame %v: %v seconds", int(frameNumber), duration)
		db.AddRenderDuration(id, int(frameNumber), duration)
	}
	return nil
}

func RemoveContainer(ctx context.Context, name string) error {
	// Remove the container after completion
	rmCmd := exec.CommandContext(ctx, "docker", "rm", name)
	err := rmCmd.Run()
	if err != nil {
		videoUpscalerLogger.Logger.Error(err.Error())
	}
	return err
}

// CountFilesInDirectory counts the number of files in a given directory
func CountFilesInDirectory(directoryPath string) int {
	// Read the directory contents
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		videoUpscalerLogger.Logger.Error(err.Error())
		return 0
	}

	// Count only files (not subdirectories)
	fileCount := 0
	for _, file := range files {
		if !file.IsDir() {
			fileCount++
		}
	}
	return fileCount
}

// FormatFrameFilename returns the correct filename for a given frame number.
func FormatFrameFilename(frameNumber int) string {
	return fmt.Sprintf("frame_%06d.png", frameNumber)
}

func isARM64() bool {
	videoUpscalerLogger.Logger.Debug("isARM64: %s", runtime.GOARCH)
	return runtime.GOARCH == "arm64"
}

func IsContainerExited(threadId string) (bool, error) {
	containerName := "upscaler-cpu" + threadId
	cmd := exec.Command(
		"docker", "ps", "-a",
		"--filter", "name="+containerName,
		"--filter", "status=exited",
		"--format", "{{.Names}}",
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, err
	}

	// Trim and check if the container name appears in output
	result := strings.TrimSpace(out.String())
	return result == containerName, nil
}
