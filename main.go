package main

import (
	"bufio"
	"strconv"
	"errors"
	"fmt"
	"os"
	"strings"
	"regexp"
	"math"
	"flag"
)

const colorReset = "\033[0m"
const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorYellow = "\033[33m"
const colorHYellow = "\033[93m"
const colorBlue = "\033[34m"
const colorPurple = "\033[35m"
const colorCyan = "\033[36m"
const colorWhite = "\033[37m"


type hunkTextMap struct {
	hunkLine	int
	textLine	int
	textStr		string
	textOpt		string
}

// Patch represents a unified diff patch.
type MapResult struct {
	LineText   string
	LineNumber int
}

// Patch represents a unified diff patch.
type Patch struct {
	Hunks      []*Hunk
	LineNumber int
}

// Hunk represents a hunk in a unified diff patch.
type Hunk struct {
	FileName   		string
	HunkNo			int
	OriginalStartLine	int
	OriginalLines    	int
	ModifiedStartLine	int
	ModifiedLines    	int
	Lines             	[]Line
	Description       	string
}

// Line represents a line in a hunk.
type Line struct {
	Operation string // " ", "+", or "-"
	Content   string
}

// ParsePatch parses a unified diff patch and returns a Patch structure.
func ParsePatch(patchContent string) (*Patch, error) {
	scanner := bufio.NewScanner(strings.NewReader(patchContent))
	patch := &Patch{}
	var currentHunk *Hunk = nil
	var FileName string
	hn := 0
	regexPattern := regexp.MustCompile(`^[ab]/`)

	for scanner.Scan() {
		line := scanner.Text()
		if (strings.HasPrefix(line, "index ")|| strings.HasPrefix(line, "+++ ")|| strings.HasPrefix(line, "--- ")) {
			continue
		}
		if strings.HasPrefix(line, "diff --git") {
			hn = 0
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				FileName = regexPattern.ReplaceAllString(parts[3], "")
			}
		} else if strings.HasPrefix(line, "@@") {
			// Parse hunk information
			hn ++
			currentHunk = &Hunk{}
			currentHunk.FileName = FileName
			currentHunk.HunkNo = hn
			patch.Hunks = append(patch.Hunks, currentHunk)
			err := parseHunkHeader(line, currentHunk)
			if err != nil {
				return nil, err
			}
		} else if currentHunk != nil {
			tmp := parseLine(line)
			if (tmp.Operation == " " || tmp.Operation == "+" || tmp.Operation == "-" ) {
				currentHunk.Lines = append(currentHunk.Lines, tmp)
			}
		}
	}

	return patch, nil
}


func parseHunkHeader(line string, hunk *Hunk) error {
	// Use regular expression to extract hunk information
	re := regexp.MustCompile(`@@ -(\d+),(\d+) \+(\d+),(\d+) @@(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 6 {
		return errors.New("invalid hunk header format")
	}

	// Extract values from regular expression matches
	originalStartLine, err := strconv.Atoi(matches[1])
	if err != nil {
		return err
	}

	originalLines, err := strconv.Atoi(matches[2])
	if err != nil {
		return err
	}

	modifiedStartLine, err := strconv.Atoi(matches[3])
	if err != nil {
		return err
	}

	modifiedLines, err := strconv.Atoi(matches[4])
	if err != nil {
		return err
	}

	hunk.Description = strings.TrimSpace(matches[5])
	hunk.OriginalStartLine = originalStartLine
	hunk.OriginalLines = originalLines
	hunk.ModifiedStartLine = modifiedStartLine
	hunk.ModifiedLines = modifiedLines

	return nil
}

func parseLine(line string) Line {
	return Line{
		Operation: line[:1],
		Content:   line[1:],
	}
}

func parseNumber(s string) (int, error) {
	return strconv.Atoi(s)
}

func readLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func mapHunk(text []string, hunk Hunk) []hunkTextMap {
        var bestRes []hunkTextMap

	positionMap := make(map[string][]int)
        bestScore :=  math.MinInt


	for i, value := range text {
		positionMap[value] = append(positionMap[value], i)
	}

        for position := 0; position < len(text); position++ {
                score, res := matchScore(positionMap, hunk.Lines, position, len(text))
                if score > bestScore {
                        bestScore = score
                        bestRes = res
                }
        }

	return bestRes
}

func matchScore(positionMap map[string][]int, hunkText []Line, position, textSize int) (int, []hunkTextMap) {
	var resMap []hunkTextMap
	currentPosition := position
	score := 0
	initialOffset := -1
	prevPos :=1

	// Iterate over hunk lines
	for i, hunkLine := range hunkText {
		// Skip lines with operation "+"
		if currentPosition+len(hunkText)>textSize {
			score = -textSize
			break
		}
		if hunkLine.Operation == "+" {
			continue
		}

		// Find the positions of the content in the text
		contentPositions, exists := positionMap[hunkLine.Content]

		if !exists {
			// No match for the current hunk line, penalize and skip
			resMap = append(resMap, hunkTextMap{i, -1, hunkLine.Content, hunkLine.Operation})
			score -= 1
			continue
		}

		// Check best position if any
		bestPos := -1
		bestPosScore := math.MaxInt
		for _, pos := range contentPositions {
			if pos >= currentPosition {
				if pos - currentPosition < bestPosScore {
					bestPos = pos
					bestPosScore = pos - currentPosition
				}
			} else {
				continue
			}
		}
		if initialOffset < 0 {
			initialOffset = bestPos
			prevPos=bestPos
		}
		// Update score based on the position check
		if bestPos<0 {
			resMap = append(resMap, hunkTextMap{i, -1, hunkLine.Content, hunkLine.Operation})
			score -= 1
		} else {
			score += len(hunkText) - bestPosScore  - (bestPos-prevPos)
			currentPosition = bestPos
			prevPos = bestPos
			resMap = append(resMap, hunkTextMap{i, bestPos, hunkLine.Content, hunkLine.Operation})
		}
	}

	return score - initialOffset, resMap
}

func longestTokenSize(line string) int {
	re := regexp.MustCompile(`\b\w+\b`)
	tokens := re.FindAllString(line, -1)
	maxLength := 0

	for _, token := range tokens {
		tokenLength := len(token)
		if tokenLength > maxLength {
			maxLength = tokenLength
		}
	}

	return maxLength
}

func ApplyPatch(patch *Patch) string {
type ResItems struct{
	Hunk int
	File string
	Commit string
	RefText string
	Reason string
}

	output := ""

	commitHashes := make(map[string]ResItems)
	textLineStats := make(map[string]int)

	for _, hunk := range patch.Hunks {
		fileLines, err := readLinesFromFile(hunk.FileName)
			if err != nil {
			panic("sdf");
		}
		for _, l := range fileLines {
			textLineStats[l]++
		}

		output += fmt.Sprintf("%sProcessing hunk %s#%d%s on file %s%s%s\n", string(colorReset), string(colorYellow), hunk.HunkNo, string(colorReset), string(colorYellow), hunk.FileName, string(colorReset))
		offs := findPosition(fileLines, *hunk)
		if offs == hunk.OriginalStartLine {
			output += fmt.Sprintf("hunk %s#%d%s applies perfectly%s\n", string(colorYellow), hunk.HunkNo, string(colorGreen), string(colorReset) )
		}
		if offs == -1 {
			output += fmt.Sprintf("hunk %s#%d%s does NOT appliy%s\n", string(colorYellow), hunk.HunkNo, string(colorRed), string(colorReset))
			commits, err := gitFetchFileHistory("/home/alessandro/src/linux/", hunk.FileName)
			if err != nil {
				fmt.Printf("/home/alessandro/src/linux/%s\n", hunk.FileName)
				panic(err)
			}
			m := mapHunk(fileLines, *hunk)
			prevTextLine := -1
			for _, v := range m {
				if v.textLine == -1 { // context text patch contains but it is not present in the patch target
					output += fmt.Sprintf("%s%s%s%s%s\n", string(colorHYellow), v.textOpt, string(colorRed), v.textStr, string(colorReset))
					for _, c := range commits {
						if (longestTokenSize(v.textStr)>5 && strings.Contains(c.Patch, v.textStr[1:])){
							commitHashes[c.Hash]=ResItems{
								Hunk: hunk.HunkNo,
								File: hunk.FileName,
								Commit: c.Hash,
								RefText: v.textStr[1:],
								Reason: "missing",
							}

						}
					}
				} else {
					if (prevTextLine!=-1 && v.textLine != prevTextLine+1) {
						for i:=prevTextLine+1; i<v.textLine; i++ {
							output += fmt.Sprintf("%s%s%s\n", string(colorYellow), "#"+fileLines[i], string(colorReset))
							for _, c := range commits {
								if (longestTokenSize(fileLines[i])>5 && strings.Contains(c.Patch, "-"+fileLines[i]) && textLineStats[fileLines[i]]<3){
									commitHashes[c.Hash]=ResItems{
										Hunk: hunk.HunkNo,
										File: hunk.FileName,
										Commit: c.Hash,
										RefText: "-"+fileLines[i],
										Reason: "other",
									}
								}
							}
						}
					}
					output += fmt.Sprintf("%s%s%s%s%s\n", string(colorHYellow), v.textOpt, string(colorGreen), v.textStr, string(colorReset))
					prevTextLine = v.textLine
				}

			}
		} else {
			output += fmt.Sprintf("hunk %s#%d%s applies %swith offset %s%d%s\n", string(colorYellow), hunk.HunkNo, string(colorGreen), string(colorReset), string(colorYellow), hunk.OriginalStartLine - offs, string(colorReset))
		}
	}
	output += fmt.Sprintf("You may want to look at these commits:%s\n", string(colorHYellow))
	if len(commitHashes)>0 {
		for _, v := range commitHashes {
			output += fmt.Sprintf("File %s:H%d Commit: %s Reason %s [%s]\n", v.File, v.Hunk, v.Commit, v.Reason, v.RefText)
		}
		output += fmt.Sprintf("%s", string(colorReset))
	}
	return output
}

func findPosition(fn []string, hunk Hunk) int {
        var app []string

        for _, l := range hunk.Lines {
                if (l.Operation == " " || l.Operation == "-") {
                        app=append(app, l.Content)
                }
        }

	for i := 0; i <= len(fn)-len(app); i++ {
		if strings.Join(fn[i:i+len(app)], "\n") == strings.Join(app, "\n") {
			return i
		}
	}
	return -1
}

func main() {
	var patchFilePath string
	flag.StringVar(&patchFilePath, "patch", "0001.diff", "Specify the patch to operate with")
	flag.Parse()

	patchContent, err := os.ReadFile(patchFilePath)
	if err != nil {
		fmt.Println("Error reading patch file:", err)
		return
	}

	patch, err := ParsePatch(string(patchContent))
	if err != nil {
		fmt.Println("Error parsing patch:", err)
		return
	}

	output := ApplyPatch(patch)
	fmt.Println(output)
}

