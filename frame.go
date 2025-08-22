package videoUpscaler

// finds an specific Frame from the Frames slice
func GetFrame(frames []*VideoUpscalerThread_Frame, filename string) *VideoUpscalerThread_Frame {
	for _, frame := range frames {
		if frame.Filename == filename {
			return frame
		}
	}
	return nil
}
