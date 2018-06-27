package logger

import (
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	ACCESS = 1 << iota // 1
	FATAL              // 2
	ERROR              // 4
	WARN               // 8
	INFO               // 16
	DEBUG              // 32
)

const (
	TIME_SYNC          = time.Second
	LOG_FLAGS          = log.LstdFlags
	DEFAULT_MAXDEPTH   = 3
	DEFAULT_STACKTRACE = true

	LEVEL_ALL        = ACCESS | FATAL | ERROR | WARN | INFO | DEBUG
	LEVEL_PRODUCTION = LEVEL_ALL ^ DEBUG
)

const (
	COMPRESS_GZIP = iota
	COMPRESS_ZIP
)

var (
	ignoreFiles        = map[string]bool{"proc.c": true, "asm_amd64.s": true}
	logTypes_EnumName  = map[int]string{ACCESS: "", FATAL: "FATAL", ERROR: "ERROR", WARN: "WARN", INFO: "INFO", DEBUG: "DEBUG"}
	logTypes_EnumValue = map[string]int{"ACCESS": ACCESS, "FATAL": FATAL, "ERROR": ERROR, "WARN": WARN, "INFO": INFO, "DEBUG": DEBUG}
	level_EnumName     = map[int]string{LEVEL_ALL: "LEVEL_ALL", LEVEL_PRODUCTION: "LEVEL_PRODUCTION"}
	level_EnumValue    = map[string]int{"LEVEL_ALL": LEVEL_ALL, "LEVEL_PRODUCTION": LEVEL_PRODUCTION}

	DEFAULT_WRITESTACKTRACE = map[int]bool{ERROR: true, WARN: true} //uppercase to "simulate" a const
)

// Struct Logger
type Logger struct {
	fileHandler     *os.File
	log             *log.Logger
	filepath        string
	logDate         time.Time
	time_sync       time.Duration
	level_log       int
	rotateFiles     bool
	writeStackTrace map[int]bool
	compressMode    int
	maxDepth        int
	dirPath         string
	filename        string
}

//Creates a new instance of logger
//fp is file with complete path. Ex: /var/log/mylog.log
//level define which log types to write to the file. Ex: logger.FATAL|logger.INFO.
// rotate define whether to rotate the log automatically or not
func New(fp string, level int, rotate bool) (l *Logger, err error) {
	if fp == "" {
		return nil, errors.New("fileName can not be empty")
	}

	l = &Logger{}

	l.filepath = fp

	l.dirPath, l.filename = filepath.Split(l.filepath)

	l.SetTimeSync(TIME_SYNC)

	l.SetLevel(level)

	l.rotateFiles = rotate
	l.setLogDate()
	l.SetMaxDepth(DEFAULT_MAXDEPTH)
	l.SetStackTrace(DEFAULT_WRITESTACKTRACE)

	l.fileHandler, err = l.createFile(l.filepath)
	if err != nil {
		return nil, err
	}

	if l.rotateFiles {
		l.SetCompressModeGzip()
		go l.rotateFile()
	}

	go l.sync()

	return
}

// Public Methods Set's

//define log types write stack trace
func (l *Logger) SetStackTrace(list map[int]bool) {
	l.writeStackTrace = list
}

// Define sync time to save log in file
func (l *Logger) SetTimeSync(t time.Duration) {
	if t > 0 {
		l.time_sync = t
	}
}

//define max depth to write stack
func (l *Logger) SetMaxDepth(m int) {
	l.maxDepth = m
}

// Define which logs types should be save in log file
func (l *Logger) SetLevel(lvl int) {
	l.level_log = lvl
}

//define gzip as compression mode
//note: works only with active file rotation
func (l *Logger) SetCompressModeGzip() {
	l.setCompressMode(COMPRESS_GZIP)
}

//define zip as compression mode
//note: works only with active file rotation
func (l *Logger) SetCompressModeZip() {
	l.setCompressMode(COMPRESS_ZIP)
}

func (l *Logger) WritePanic(rec interface{}, stack []byte) {
	txt := "========================== panic ==========================\n"
	txt += fmt.Sprintf("%v\n", rec)
	txt += fmt.Sprintf("%s\n", stack)
	txt += "==================================================================="
	l.Printf(ERROR, "%s", txt)
}

// Close log file
func (l *Logger) Close() {
	l.fileHandler.Close()
}

func (l *Logger) GetTypeString(typeLog int) string {
	return logTypes_EnumName[typeLog]
}

// typeLog is a level log. Ex:INFO, ERROR, WARN
// txt is a message to write
func (l *Logger) Print(typeLog int, txt ...interface{}) {
	if !l.checkWrite(typeLog) {
		return
	}
	str := parsePrint(txt...)
	l.checkPrintStack(typeLog, &str)

	l.log.Print(l.GetTypeString(typeLog), str)
	l.isFatal(typeLog)
}

// typeLog is a level log. Ex:INFO, ERROR, WARN
// txt is a message to write
func (l *Logger) Println(typeLog int, txt ...interface{}) {
	if !l.checkWrite(typeLog) {
		return
	}

	str := parsePrint(txt...)
	l.checkPrintStack(typeLog, &str)
	l.log.Println(l.GetTypeString(typeLog), str)
	l.isFatal(typeLog)
}

// typeLog is a level log. Ex:INFO, ERROR, WARN
// format is text format
// txt is variables that will be formatted
func (l *Logger) Printf(typeLog int, format string, txt ...interface{}) {
	if !l.checkWrite(typeLog) {
		return
	}
	l.checkPrintStack(typeLog, &format)
	l.log.Printf(l.GetTypeString(typeLog)+" "+format, txt...)
	l.isFatal(typeLog)
}

// Same as the Print function but after writing to the file, it kills the application with an Exit (1)
func (l *Logger) Fatal(txt ...interface{}) {
	typeLog := FATAL
	if !l.checkWrite(typeLog) {
		return
	}
	str := parsePrint(txt...)
	l.checkPrintStack(typeLog, &str)
	l.log.Print(l.GetTypeString(typeLog), str)
	//	l.fileHandler.Close()
	//	os.Exit(1)
	//l.closeAndKill()
	l.isFatal(typeLog)
}

// Same as the Printf function but after writing to the file, it kills the application with an Exit (1)
func (l *Logger) Fatalf(format string, txt ...interface{}) {
	typeLog := FATAL
	if !l.checkWrite(typeLog) {
		return
	}
	l.checkPrintStack(typeLog, &format)
	l.log.Printf(l.GetTypeString(typeLog)+" "+format, txt...)
	//	l.fileHandler.Close()
	//	os.Exit(1)
	//l.closeAndKill()
	l.isFatal(typeLog)
}

// Same as the Println function but after writing to the file, it kills the application with an Exit (1)
func (l *Logger) Fatalln(txt ...interface{}) {
	typeLog := FATAL
	if !l.checkWrite(typeLog) {
		return
	}

	str := parsePrint(txt...)
	l.checkPrintStack(typeLog, &str)
	l.log.Println(l.GetTypeString(typeLog), str)
	//l.fileHandler.Close()
	//os.Exit(1)
	//l.closeAndKill()
	l.isFatal(typeLog)
}

// Private methods

// save log in disk
func (l *Logger) sync() {
	for {

		if _, er := os.Stat(l.filepath); er != nil {
			var err error
			l.fileHandler, err = l.createFile(l.filepath)
			if err != nil {
				fmt.Printf("Failed rotate log. Error [%s]", err.Error())
				os.Exit(1)
			}
		}

		l.fileHandler.Sync() //write in disk

		time.Sleep(l.time_sync)
	}
}

func (l *Logger) moveFiles() (err error) {
	defer l.setLogDate()

	//only move and compress if size > 0.
	if fi, _ := os.Stat(l.filepath); fi.Size() > 0 {

		path, filename := l.dirPath, l.filename
		//convert log date to string
		dateParse := l.logDate.Format("20060102")

		//build a new file name
		if n := strings.LastIndex(filename, "."); n != -1 {
			filename = filename[:n] + "_" + dateParse + filename[n:]
		}
		// join dir and new file
		dest := filepath.Join(path, filename)

		err = os.Rename(l.filepath, dest)
		if err != nil {
			return
		}

		//make a compression
		err = l.execCompressFile(dest)
		if err != nil {
			return
		}
	}

	if err != nil {
		return
	}
	return
}
func (l *Logger) isFatal(t int) {
	if t != FATAL {
		return
	}
	l.closeAndKill()
}
func (l *Logger) closeAndKill() {
	l.fileHandler.Close()
	os.Exit(1)
}

//rotate log file, always that alter day
func (l *Logger) rotateFile() {
	date := time.Now()
	currentDate := date
	for {
		date = time.Now()
		//move and rotate files, if alter day
		if date.Day() != currentDate.Day() {
			l.moveFiles()
			currentDate = time.Now()
		}
		//TODO: calc sleep next day
		time.Sleep(time.Second)
	}
}

func (l *Logger) whoPrintStack() string {
	pc, file, line, ok := runtime.Caller(l.maxDepth)
	if !ok {
		return "[unknown - 0] "
	}
	me := runtime.FuncForPC(pc)
	if me == nil {
		return "[- - 0 ] "
	}
	path := strings.Split(file, "/")
	functioName := strings.Split(me.Name(), ".")
	return " [" + functioName[len(functioName)-1] + " " + path[len(path)-1] + " " + fmt.Sprintf("%v", line) + "] "
}

//check log level write stack
func (l *Logger) checkPrintStack(typeLog int, str *string) {
	if _, ok := l.writeStackTrace[typeLog]; ok {
		*str = l.whoPrintStack() + *str
	}
}

//creates the log file, if the directory does not exist it will be created
func (l *Logger) createFile(fl string) (*os.File, error) {
	err := makeDir(fl)
	if err != nil {
		return nil, err
	}

	fh, err := os.OpenFile(fl, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0766)
	if err != nil {
		return nil, err
	}

	l.log = log.New(fh, "", log.LstdFlags)

	return fh, nil
}

func (l *Logger) checkWrite(lvl int) bool {
	return l.level_log&lvl != 0
}

func (l *Logger) setCompressMode(mode int) {
	l.compressMode = mode
}

func (l *Logger) setLogDate() {
	l.logDate = time.Now()
}

func (l *Logger) execCompressFile(oldFile string) error {

	switch l.compressMode {
	case COMPRESS_GZIP:
		if err := l.compressGzip(&oldFile); err != nil {
			return err
		}
	case COMPRESS_ZIP:
		if err := l.compressZip(&oldFile); err != nil {
			return err
		}

	}
	os.Remove(oldFile)
	return nil
}
func (l *Logger) compressGzip(oldFile *string) error {
	content, err := ioutil.ReadFile(*oldFile)
	if err != nil {
		return err
	}
	f, err := os.Create(*oldFile + ".gz")
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		return err
	}
	w.ModTime = time.Now()
	w.Write(content)
	w.Close()
	return nil
}
func (l *Logger) compressZip(oldFile *string) error {
	fileZip, err := os.Create(*oldFile + ".zip")
	if err != nil {
		return err
	}
	defer fileZip.Close()

	// Create a new zip archive.
	w := zip.NewWriter(fileZip)
	defer w.Close()

	zipfile, err := os.Open(*oldFile)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	// Get the file information
	info, err := zipfile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Modified = time.Now()
	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, zipfile)

	return err

}

func parsePrint(r ...interface{}) (v string) {
	for e := range r {
		v += fmt.Sprintf("%v", r[e])
	}
	return
}

// create a dir struct
func makeDir(fn string) error {
	dir := filepath.Dir(fn)

	switch dir {
	//If the path is empty, Dir returns "."
	case ".":
		return nil
	case "/": // ex: /name.log, Dir returns	/
	default:
		dir += "/"
	}

	return os.MkdirAll(dir, 0766)
}
