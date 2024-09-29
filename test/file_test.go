package test

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRewriteIDs(t *testing.T) {
	file, err := os.Open("/Users/jujiajia/playground/lightning/test-t-250M-2/test.t.1.csv")
	require.NoError(t, err)
	defer file.Close()
	outFile, err := os.Create("/Users/jujiajia/playground/lightning/test-t-250M-2/test.t.1.csv-out")
	require.NoError(t, err)
	defer outFile.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		splits := strings.SplitN(text, ",", 2)
		id, err := strconv.Atoi(splits[0])
		require.NoError(t, err)
		id += 330000
		_, err = outFile.WriteString(strconv.Itoa(id) + "," + splits[1] + "\n")
		require.NoError(t, err)
	}
}
