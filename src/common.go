package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Error struct {
	Message string
	Err     error
}

func (e Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf(">>> Error : %v : %v", e.Message, e.Err)
	}
	return fmt.Sprintf(">>> Error : %v.", e.Message)
}

// Splits the episode into multiple media files that we will later merge together
func Split(fileName string) error {
	// TODO check if file exists before attempting extraction
	path, err := exec.LookPath("engine\\flvextract.exe")
	if err != nil {
		return Error{"Unable to find flvextract.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path, "-v", "-a", "-t", "-o", "temp\\"+fileName+".flv")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		return Error{"There was an error while executing our extracter", err}
	}
	err = cmd.Wait()
	if err != nil {
		return Error{"There was an error while extracting", err}
	}
	return nil
}

// Merges all media files including subs
func Merge(fileName string) error {
	// TODO check if files exist before attempting final merge
	path, err := exec.LookPath("engine\\mkvmerge.exe")
	if err != nil {
		return Error{"Unable to find mkvmerge.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to split our flv
	cmd := exec.Command(path,
		"-o", "temp\\"+fileName+".mkv",
		"--language", "0:eng",
		"temp\\"+fileName+".ass",
		"temp\\"+fileName+".264",
		"--aac-is-sbr", "0",
		"temp\\"+fileName+".aac")

	// Executes the extraction and waits for a response
	err = cmd.Start()
	if err != nil {
		return Error{"There was an error while executing our merger", err}
	}

	// Waits for the merge to complete
	err = cmd.Wait()
	if err != nil {
		return Error{"There was an error while merging", err}
	}

	_, err = exec.LookPath("temp\\" + fileName + ".mkv")
	if err != nil {
		return Error{"Merged MKV was not found after merger", err}
	}
	// Erases all old media files that we no longer need
	os.Remove("temp\\" + fileName + ".ass")
	os.Remove("temp\\" + fileName + ".264")
	os.Remove("temp\\" + fileName + ".txt")
	os.Remove("temp\\" + fileName + ".aac")
	os.Remove("temp\\" + fileName + ".flv")
	return nil
}

// Cleans up the mkv, optimizing it for playback as well as old remaining files
func Clean(fileName string) error {
	// TODO check if file exists before attempting extraction
	path, err := exec.LookPath("engine\\mkclean.exe")
	if err != nil {
		return Error{"Unable to find mkclean.exe in \\engine\\ directory", err}
	}

	// Creates the command which we will use to clean our mkv to "video.clean.mkv"
	cmd := exec.Command(path, "--optimize", "temp\\"+fileName+".mkv")

	// Executes the cleaning and waits for a response
	err = cmd.Start()
	if err != nil {
		return Error{"There was an error while executing our mkv optimizer", err}
	}
	err = cmd.Wait()
	if err != nil {
		return Error{"There was an error while optimizing our mkv", err}
	}

	// Deletes the old, un-needed dirty mkv file
	os.Remove("temp\\" + fileName + ".mkv")
	os.Rename("temp\\clean."+fileName+".mkv", "temp\\"+fileName+".mkv")
	return nil
}

// Gets user input from the user and unmarshalls it into the input
func getStandardUserInput(prefixText string, input *string) error {
	fmt.Printf(prefixText)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		*input = scanner.Text()
		break
	}
	if err := scanner.Err(); err != nil {
		return Error{"There was an error getting standard user input", err}
	}
	return nil
}

// Constructs an episode file name and returns the file name cleaned
func generateEpisodeFileName(showTitle string, seasonNumber int, episodeNumber float64, description string) string {
	// Pads season number with a 0 if it's less than 10
	seasonNumberString := strconv.Itoa(seasonNumber)
	if seasonNumber < 10 {
		seasonNumberString = "0" + strconv.Itoa(seasonNumber)
	}

	// Pads episode number with a 0 if it's less than 10
	episodeNumberString := strconv.FormatFloat(episodeNumber, 'f', -1, 64)
	if episodeNumber < 10 {
		episodeNumberString = "0" + strconv.FormatFloat(episodeNumber, 'f', -1, 64)
	}

	// Constructs episode file name and returns it
	fileName := strings.Title(showTitle) + " - S" + seasonNumberString + "E" + episodeNumberString + " - " + description
	return cleanFileName(fileName)
}

// Cleans the new file/folder name so there won't be any write issues
func cleanFileName(fileName string) string {
	newFileName := fileName // Strips out any illegal characters and returns our new file name
	for _, illegalChar := range []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"} {
		newFileName = strings.Replace(newFileName, illegalChar, " ", -1)
	}
	return newFileName
}
