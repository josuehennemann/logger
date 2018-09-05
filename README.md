# logger
Package to write log files

#### Main usage:
```golang
Log, _:= logger.New("/path/of/file.log", logger.LEVEL_ALL, true)
Log.Printf(logger.INFO, "Text formatted to write in log string (%s) int (%d)", "value", 90)
Log.Println(logger.INFO, "Text to write in log")
```
This package implements the following log types: **ACCESS, FATAL, ERROR, WARN, INFO, DEBUG**.

Right after a FATAL log is written, the application will be killed with a Exit (1)

To initialize logger its necessary to define which types can be written. This package has 2 log levels.

**LEVEL_ALL**: Writes all log types

**LEVEL_PRODUCTION**: Writes all log types except DEBUG

This allows you to initialize the log according to the environment in which your application is running (eg development and production)

#### Custom Log Levels:

A custom level can be defined using bitwise. Example:
```
Log, _ := logger.New("/path/of/file.log", logger.INFO|logger.WARN, true)
```
In this case, only INFO and WARN logs will be written.
```
Log.Println(logger.INFO, "Text info")
Log.Println(logger.WARN, "Text warn")
Log.Println(logger.ERROR, "Text error") // this won't be logged
```
#### File rotation:

At the end of the day, log files will be compressed and a new empty log file will be created.
Compression is done using gzip, but can be changed at any time to zip.
To change the compression method, use the functions:
```
Log.SetCompressModeGzip()
Log.SetCompressModeZip()
```
#### Stack:

By default, ERROR and WARN logs write a stack stating where the log was written. Example:
```
func testWriteLog() {
  Log.Printf(ERROR, "test write stack")
}
```
Will log:
2018/06/20 22:42:24 ERROR [testWriteLog main.go 19] test write stack

To change this setting use the function
```
Log.SetStackTrace(logger.WARN|logger.INFO)
```

### Contributors
[Christopher Madalosso Burin](https://github.com/chriscmb)
