/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stracciato",
	Short: "Analyse strace outputs",
	Long: `The tool will analyse a strace output that was taken 
	with at least some mandatory options. Generally using 
	strace with "-fttTvxy" is recommended.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd, args)
	},
}

type stats struct {
	count      int64
	sum        float64
	average    float64
	min        float64
	max        float64
	unfinished int64
	unknown    []string
}

var (
	syscalls = make(map[string]stats)
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Display additional messages")

}

func sum(numbers []float64) float64 {
	sum := 0.0
	for _, num := range numbers {
		sum += num
	}
	return sum
}

func updateStats(syscall string, timingFloat float64) {
	tempStat := syscalls[syscall]
	tempStat.count++
	tempStat.sum += timingFloat
	tempStat.average = tempStat.sum / float64(tempStat.count)
	if timingFloat > tempStat.max {
		tempStat.max = timingFloat
	}
	if timingFloat < tempStat.min || tempStat.min == 0 {
		tempStat.min = timingFloat
	}
	syscalls[syscall] = tempStat
}

func processThread(wg *sync.WaitGroup, entries []string) error {
	defer wg.Done()
	//fmt.Println("============\n", entries)
	syscallRegex, err := regexp.Compile("(?P<syscall>[a-zA-Z_]+)\\(")
	if err != nil {
		return err
	}
	syscallUnfRegex, err := regexp.Compile("(?P<syscallUn>[a-zA-Z_]+)")
	if err != nil {
		return err
	}
	timedRegex, err := regexp.Compile("[0-9]+(\\.[0-9]+)*")
	if err != nil {
		return err
	}
	unfinishedRegex, err := regexp.Compile(".*unfinished.*")
	if err != nil {
		return err
	}
	for i := 0; i < len(entries); i++ {
		// The second field should always contain a syscall
		// Add them to a global map if the match a regex
		firstLine := entries[i]
		lineArr := strings.Fields(firstLine)
		//fmt.Println("\t--", firstLine)
		//fmt.Println("\t +contains a syscall? ", syscallRegex.MatchString(lineArr[2]))
		if syscallRegex.MatchString(lineArr[2]) {
			syscallGroupRegex := syscallRegex.FindStringSubmatch(lineArr[2])
			syscall := syscallGroupRegex[1]
			timing := lineArr[len(lineArr)-1]
			//fmt.Printf("%s: %s\n", syscall, timing)
			if timedRegex.MatchString(string(timing)) {
				// We got a completed syscall, add it to global map
				var replacer = strings.NewReplacer("<", "", ">", "")
				timingTrimmmed := replacer.Replace(string(timing))
				timingFloat, err := strconv.ParseFloat(timingTrimmmed, 4)
				if err != nil {
					fmt.Println("Error converting timing to float: ", err)
					return err
				}
				//fmt.Printf("\t +Line contains a syscall: %s with timing %.6f\n", syscall, timingFloat)
				updateStats(syscall, timingFloat)

			} else {
				// It's a syscall but it's unfinished
				//fmt.Println("No completed syscall")
				if unfinishedRegex.MatchString(lineArr[len(lineArr)-2]) {
					// Get the next line
					secondLine := entries[i+1]
					//fmt.Println(firstLine)
					//fmt.Println(secondLine)
					lineArr2 := strings.Fields(secondLine)
					timing := lineArr2[len(lineArr2)-1]
					//fmt.Println(lineArr2)
					syscall2GroupRegex := syscallUnfRegex.FindStringSubmatch(lineArr2[3])
					//fmt.Printf("%s == %s\n", syscall, syscall2GroupRegex[1])
					if timedRegex.MatchString(string(timing)) && (syscall2GroupRegex[1] == syscall) {
						var replacer = strings.NewReplacer("<", "", ">", "")
						timingTrimmmed := replacer.Replace(string(timing))
						timingFloat, err := strconv.ParseFloat(timingTrimmmed, 4)
						if err != nil {
							fmt.Println("Error converting timing to float: ", err)
							return err
						}
						//fmt.Printf("\t +Line contains a syscall: %s with timing %.6f\n", syscall, timingFloat)
						updateStats(syscall, timingFloat)
					}
					i++
				} else {
					// It's a syscall but for some reasons it doesn't have timings
					// or unfinished. Save it for now
					tempStat := syscalls[syscall]
					tempStat.unknown = append(tempStat.unknown, firstLine)
					syscalls[syscall] = tempStat
				}
			}
		}

	}
	//fmt.Println(syscalls)
	return nil
}

func run(cmd *cobra.Command, args []string) error {
	var wg sync.WaitGroup
	var unknownCalls = false
	file, err := os.Open(args[0])
	if err != nil {
		log.Fatal("Error opening the file", args[0], err)
	}
	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatal("Error retrieving file info:", err)
	}
	if fileinfo.IsDir() {
		log.Fatal("This can work only on files, not directories")
	}
	threads := make(map[string][]string)
	scanner := bufio.NewScanner(file)
	// Let's store each thread in a separate map entry
	for scanner.Scan() {
		line := scanner.Text()
		threads[strings.Fields(line)[0]] = append(threads[strings.Fields(line)[0]], line)
	}
	for _, entries := range threads {
		wg.Add(1)
		processThread(&wg, entries)

	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading the file: ", err)
	}
	fmt.Printf("%-20s %-10s %-10s %-10s %-10s %-10s\n", "Syscall", "Count", "Average", "Min", "Max", "Total")
	for sysc, alltimings := range syscalls {
		fmt.Printf("%-20s %-10d %-10f %-10f %-10f %-10f\n", sysc, alltimings.count, alltimings.average, alltimings.min, alltimings.max, alltimings.sum)
		if len(alltimings.unknown) > 0 {
			unknownCalls = true
		}
	}
	fmt.Printf("\nNumber of threads: %d\n", len(threads))
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		log.Fatal("Error retrieving verbose value: ", err)
	}
	if unknownCalls {
		fmt.Printf("\nThere were some unknown calls")
		if !verbose {
			fmt.Printf(", use '--verbose/-v' to see them.\n")
		} else {
			fmt.Printf(":\n")
			for sysc, alltimings := range syscalls {
				if len(alltimings.unknown) > 0 {
					fmt.Printf("- %s\n", sysc)
					for _, v := range alltimings.unknown {
						fmt.Printf("\t %s\n", v)
					}
				}
			}
		}
	}
	wg.Wait()
	return nil
}
