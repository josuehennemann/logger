package logger

import (
	"testing"
	"time"
	"os"
)

func TestLoggerName(t *testing.T) {

	_, err := New("", LEVEL_ALL, false)
	if err == nil {
		t.Error("fileName can not be empty")
	}
}

func TestLoggerWithRotate(t *testing.T) {
	filename := "test.log"

	logFile, err := New(filename, LEVEL_ALL, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}
	defer logFile.Close()
	logFile.Printf(ACCESS, "teste ACCESS")
	//logFile.Printf(FATAL, "teste FATAL")
	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(WARN, "teste WARN")
	logFile.Printf(INFO, "teste INFO")
	logFile.Printf(DEBUG, "teste DEBUG")
	//alter stack

	logFile.Printf(INFO, "Alter types write stack")

	logFile.SetStackTrace(ACCESS|INFO)

	logFile.Printf(ACCESS, "teste ACCESS")
	//logFile.Printf(FATAL, "teste FATAL")
	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(WARN, "teste WARN")
	logFile.Printf(INFO, "teste INFO")
	logFile.Printf(DEBUG, "teste DEBUG")

	deleteFile(filename,t)
}

func TestLoggerCompressGzip(t *testing.T) {
	filename := "test_gzip"	
	fileExt := ".log"	
	logFile, err := New(filename+fileExt, LEVEL_ALL, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}
	defer logFile.Close()
	logFile.SetCompressModeGzip()
	logFile.Printf(INFO, "Test compress gzip")

	if err := logFile.moveFiles(); err != nil {
		t.Error("Failed compress log in mode gzip", "Error", err)
	}
	//sleep by 10s, to exec func sync() to create a new file
	time.Sleep(time.Second * 10)
	deleteFile(filename+fileExt,t)
	deleteFile(filename+"_"+time.Now().Format("20060102")+fileExt+".gz",t)

}
func TestLoggerCompressZip(t *testing.T) {
	filename := "test_zip"
	fileExt := ".log"	

	logFile, err := New(filename+fileExt, LEVEL_ALL, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}
	defer logFile.Close()
	logFile.SetCompressModeZip()
	logFile.Printf(INFO, "Test compress zip")

	if err := logFile.moveFiles(); err != nil {
		t.Error("Failed compress log in mode zip", "Error", err)
	}
	//sleep by 10s, to exec func sync() to create a new file
	time.Sleep(time.Second * 10)
	deleteFile(filename+fileExt,t)
	deleteFile(filename+"_"+time.Now().Format("20060102")+fileExt+".zip",t)
}

func TestLoggerPersonalizeLevel(t *testing.T) {
	filename := "test_personalizeLevel.log"	
	logFile, err := New(filename, ERROR|INFO, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}

	defer logFile.Close()
	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(INFO, "teste INFO")

	logFile.Printf(WARN, "teste WARN")   //not write in file
	logFile.Printf(DEBUG, "teste DEBUG") //not write in file

	deleteFile(filename,t)
}

func TestLoggerProductionLevel(t *testing.T) {
	filename := "test_productionlevel.log"
	logFile, err := New(filename, LEVEL_PRODUCTION, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}

	defer logFile.Close()

	logFile.Printf(ACCESS, "teste ACCESS")

	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(WARN, "teste WARN")
	logFile.Printf(INFO, "teste INFO")
	logFile.Printf(DEBUG, "teste DEBUG") //not write in file
	deleteFile(filename,t)
}

func testStackLogger_1(l *Logger) {
	l.Printf(ERROR, "Test stack")
	testStackLogger_2(l)
}

func testStackLogger_2(l *Logger) {
	l.Printf(ERROR, "Test stack 2")
}

func deleteFile(filename string, t *testing.T){
	if err := os.Remove(filename); err != nil {
		t.Error("Failed delete file ", filename, err.Error())
	}
}
