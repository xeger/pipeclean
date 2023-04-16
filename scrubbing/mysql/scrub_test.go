package mysql_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/xeger/pipeclean/scrubbing"
	"github.com/xeger/pipeclean/scrubbing/mysql"
)

func read(t *testing.T, name string) string {
	data, err := ioutil.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %s", name, err)
	}
	return string(data)
}

func scrub(input string) string {
	reader := bufio.NewReader(bytes.NewBufferString(input))
	in := make(chan string)

	out := make(chan string)
	output := bytes.NewBuffer(make([]byte, 0, len(input)))
	writer := bufio.NewWriter(output)

	scrubber := scrubbing.NewScrubber("", nil, 0.95)
	go mysql.ScrubChan(scrubber, in, out)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		in <- line
		scrubbed := <-out
		writer.WriteString(scrubbed)
	}
	close(in)
	close(out)

	writer.Flush()
	outputString := output.String()
	fmt.Printf("----BEGIN SCRUB OUTPUT----\n%s\n----END SCRUB OUTPUT----\n", outputString)
	return outputString
}

func TestCreateTables(t *testing.T) {
	input := read(t, "create_tables.sql")
	output := scrub(input)

	if strings.Index(output, "DROP TABLE IF EXISTS") < 0 {
		t.Errorf("DROP TABLE statement is missing")
	}
	if strings.Index(output, "CREATE TABLE") < 0 {
		t.Errorf("CREATE TABLE statement is missing")
	}
}

func TestInsertNamed(t *testing.T) {
	input := read(t, "insert-named.sql")
	output := scrub(input)

	if strings.Index(output, "LOCK TABLES") < 0 {
		t.Errorf("LOCK TABLES statement is missing")
	}
	if strings.Index(output, "INSERT INTO `bank_accounts` (`id`,`routing_number`) VALUES (1,'111000025'),(2,'226073523');") < 0 {
		t.Errorf("INSERT statement not properly sanitized")
	}
	if strings.Index(output, "UNLOCK TABLES") < 0 {
		t.Errorf("UNLOCK TABLES statement is missing")
	}
}

func TestInsertPositional(t *testing.T) {
	input := read(t, "insert-positional.sql")
	output := scrub(input)

	if strings.Index(output, "LOCK TABLES") < 0 {
		t.Errorf("LOCK TABLES statement is missing")
	}
	if strings.Index(output, "INSERT INTO `emails` VALUES (1,'t@hjyemwg.com'),(2,'p@hjyemwg.com');") < 0 {
		t.Errorf("INSERT statement not properly sanitized:\n----BEGIN SQL----\n%s\n----END SQL----\n", output)
	}
	if strings.Index(output, "UNLOCK TABLES") < 0 {
		t.Errorf("UNLOCK TABLES statement is missing")
	}
}
