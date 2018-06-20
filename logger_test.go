package logger

import (
	"testing"
	"time"
)

func TestLoggerName(t *testing.T) {

	_, err := New("", LEVEL_ALL, false)
	if err == nil {
		t.Error("fileName can not be empty")
	}
}

func TestLoggerWithRotate(t *testing.T) {
	logFile, err := New("test.log", LEVEL_ALL, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}
	defer logFile.Close()
	logFile.Printf(ACCESS, "teste ACCESS")
	logFile.Printf(FATAL, "teste FATAL")
	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(WARN, "teste WARN")
	logFile.Printf(INFO, "teste INFO")
	logFile.Printf(DEBUG, "teste DEBUG")
	//alter stack

	logFile.Printf(INFO, "Alter types write stack")

	logFile.SetStackTrace(map[int]bool{ACCESS: true, INFO: true})

	logFile.Printf(ACCESS, "teste ACCESS")
	logFile.Printf(FATAL, "teste FATAL")
	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(WARN, "teste WARN")
	logFile.Printf(INFO, "teste INFO")
	logFile.Printf(DEBUG, "teste DEBUG")

}

func TestLoggerCompressGzip(t *testing.T) {
	logFile, err := New("test_gzip.log", LEVEL_ALL, true)
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
}
func TestLoggerCompressZip(t *testing.T) {
	logFile, err := New("test_zip.log", LEVEL_ALL, true)
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
}

func TestLoggerMaxDepth(t *testing.T) {
	logFile, err := New("test_gzip.log", LEVEL_ALL, true)
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
}

func TestLoggerPersonalizeLevel(t *testing.T) {
	logFile, err := New("test_personalizeLevel.log", ERROR|INFO, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}

	defer logFile.Close()
	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(INFO, "teste INFO")

	logFile.Printf(WARN, "teste WARN")   //not write in file
	logFile.Printf(DEBUG, "teste DEBUG") //not write in file

}

func TestLoggerProductionLevel(t *testing.T) {
	logFile, err := New("test_productionlevel.log", LEVEL_PRODUCTION, true)
	if err != nil {
		t.Error("failed create file", err.Error())
	}

	defer logFile.Close()

	logFile.Printf(ACCESS, "teste ACCESS")
	logFile.Printf(FATAL, "teste FATAL")
	logFile.Printf(ERROR, "teste ERROR")
	logFile.Printf(WARN, "teste WARN")
	logFile.Printf(INFO, "teste INFO")
	logFile.Printf(DEBUG, "teste DEBUG") //not write in file

}

func testStackLogger_1(l *Logger) {
	l.Printf(ERROR, "Test stack")
	testStackLogger_2(l)
}

func testStackLogger_2(l *Logger) {
	l.Printf(ERROR, "Test stack 2")
}
