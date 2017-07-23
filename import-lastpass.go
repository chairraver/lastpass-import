package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mkideal/cli"
	clix "github.com/mkideal/cli/ext"
	"github.com/pkg/errors"
)

const passCmd = "gopass"

type argT struct {
	cli.Helper
	Content      clix.File `cli:"i,file" usage:"read content from file or stdin"`
	DefaultGroup string    `cli:"d,default" usage:"name for uncategorized entries"`
	Force        bool      `cli:"f,force" usage:"force existing entries to be overwritten"`
}

var forceFlag = false
var defaultGroup = ""

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		if argv.Force {
			forceFlag = true
		}

		if argv.DefaultGroup != "" {
			defaultGroup = argv.DefaultGroup
		}

		if len(argv.Content.String()) > 0 {
			if err := process(argv.Content.String()); err != nil {
				log.Fatalln(err)
			}
		}
		return nil
	})
}

type lasspassEntry struct {
	url      string
	username string
	password string
	extra    string
	name     string
	grouping string
	fav      string
}

/* From the Ruby lasspass import version:

"#{@password}\n---\n"
"#{@grouping} / " unless @grouping.empty?
"#{@name}\n"
"username: #{@username}\n" unless @username.empty?
"password: #{@password}\n" unless @password.empty?
"url: #{@url}\n" unless @url == "http://sn"
"#{@extra}\n" unless @extra.nil?
*/

func (lpe lasspassEntry) String() string {
	str := lpe.password + "\n---\n"
	if lpe.grouping != "" {
		str += lpe.grouping + " / "
	}
	str += lpe.name + "\n"
	if lpe.username != "" {
		str += "username: " + lpe.username + "\n"
	}
	if lpe.password != "" {
		str += "password: " + lpe.password + "\n"
	}
	if lpe.url != "http://sn" {
		str += "url: " + lpe.url + "\n"
	}
	if lpe.extra != "" {
		str += lpe.extra + "\n"
	}
	return str
}

func getName(lpe lasspassEntry) string {
	// s << @grouping + "/" unless @grouping.empty?
	// s << @name unless @name == nil
	// s.gsub(/ /, "_").gsub(/'/, "")
	str := ""
	if lpe.grouping != "" {
		lpe.grouping = strings.Replace(lpe.grouping, "\\", "-", -1)
		str += lpe.grouping + "/"
	}
	if lpe.name != "" {
		str += lpe.name
	}
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, "'", "", -1)
	return str
}

func process(lp string) error {

	var lpe lasspassEntry

	passDir, err := getPassDir()
	if err != nil {
		return errors.Wrap(err, "getting data directory")
	}

	entries := 0
	errors := 0

	r := csv.NewReader(strings.NewReader(lp))
	for {
		// url,username,password,extra,name,grouping,fav
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if record[0] == "url" {
			continue
		}

		entries++

		lpe.url = record[0]
		lpe.username = record[1]
		lpe.password = record[2]
		lpe.extra = record[3]
		lpe.name = strings.Replace(record[4], "/", "-", -1)
		lpe.grouping = record[5]
		if record[5] == "" {
			lpe.grouping = defaultGroup
		}
		lpe.fav = record[6]

		err = doImport(lpe, passDir)
		if err != nil {
			log.Println("import error: " + err.Error())
			errors++
		}
	}

	log.Printf("# of entries handled: %d\n", entries)
	log.Printf("# of errors: %d\n", errors)

	return nil
}

func doImport(lpe lasspassEntry, passDir string) error {

	entryName := getName(lpe)

	_, err := os.Stat(passDir + string(os.PathSeparator) + entryName + ".gpg")

	if err == nil && !forceFlag {
		return errors.New("existing entry " + entryName + " will not be overwritten")
	} else if pe, ok := err.(*os.PathError); ok && !os.IsNotExist(pe) {
		return errors.Wrap(err, "ignoring existing pass entry")
	}

	log.Printf("Creating record '%s'...\n", entryName)

	cmd := exec.Command(passCmd, "insert", "-m", entryName)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, fmt.Sprint(lpe))
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	log.Printf("%s\n", out)

	return nil
}

func getPassDir() (string, error) {

	cmd := exec.Command(passCmd, "config", "path")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	res, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	path := strings.Replace(string(res), "path: ", "", -1)
	path = strings.Replace(path, "\n", "", -1)

	log.Println(path)

	return path, nil
}
