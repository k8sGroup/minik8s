package klog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

var output io.Writer

var debug = true
var pathName = "./klog.log"

const prefixFmt string = "[%s]\t%s - %d %s "

func init() {
	if pathName == "" {
		output = os.Stderr
		return
	}
	logFile, err := os.OpenFile(pathName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("open log file failed, output to stderr\n")
		output = os.Stderr
	}
	output = logFile
}

// Infof outputs log with level Info
func Infof(f string, v ...any) {
	funcName, file, line, _ := runtime.Caller(1)
	strBuilder := strings.Builder{}
	strBuilder.WriteString(prefixFmt)
	strBuilder.WriteString(f)
	var a = []any{"Info", file, line, runtime.FuncForPC(funcName).Name()}
	a = append(a, v...)
	_, _ = fmt.Fprintf(output, strBuilder.String(), a...)
}

// Warnf outputs log with level Warn
func Warnf(f string, v ...any) {
	funcName, file, line, _ := runtime.Caller(1)
	strBuilder := strings.Builder{}
	strBuilder.WriteString(prefixFmt)
	strBuilder.WriteString(f)
	var a = []any{"Warn", file, line, runtime.FuncForPC(funcName).Name()}
	a = append(a, v...)
	_, _ = fmt.Fprintf(output, strBuilder.String(), a...)
}

// Fatalf output log and the program exits with code 1
func Fatalf(f string, v ...any) {
	funcName, file, line, _ := runtime.Caller(1)
	strBuilder := strings.Builder{}
	strBuilder.WriteString(prefixFmt)
	strBuilder.WriteString(f)
	var a = []any{"Fatal", file, line, runtime.FuncForPC(funcName).Name()}
	a = append(a, v...)
	_, _ = fmt.Fprintf(output, strBuilder.String(), a...)
	os.Exit(1)
}

// Errorf outputs log with level Error
func Errorf(f string, v ...any) {
	funcName, file, line, _ := runtime.Caller(1)
	strBuilder := strings.Builder{}
	strBuilder.WriteString(prefixFmt)
	strBuilder.WriteString(f)
	var a = []any{"Error", file, line, runtime.FuncForPC(funcName).Name()}
	a = append(a, v...)
	_, _ = fmt.Fprintf(output, strBuilder.String(), a...)
}

/*
Debugf outputs log with level Debug.

Set debug false in release
*/
func Debugf(f string, v ...any) {
	if !debug {
		return
	}
	funcName, file, line, _ := runtime.Caller(1)
	strBuilder := strings.Builder{}
	strBuilder.WriteString(prefixFmt)
	strBuilder.WriteString(f)
	var a = []any{"Debug", file, line, runtime.FuncForPC(funcName).Name()}
	a = append(a, v...)
	_, _ = fmt.Fprintf(output, strBuilder.String(), a...)
}
