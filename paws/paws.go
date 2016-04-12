package paws

import (
	//"fmt"
	"github.com/joho/godotenv"
	"os"
)

const (
	TABLE_INFO_ROW  = iota // 0
	IMPORT_INFO_ROW        // 1
	KEY_INFO_ROW           // 2
	ALIAS_INFO_ROW         // 3
	FIELD_INFO_ROW         // 4
	START_DATA_ROW         // 5
)

type Column struct {
	Name, Alias string
	Key, Import bool
	Index       int
	SqlType     string
	IsRTF       bool
}

type Data struct {
	Value string
	Index int
}

func GetPAWSInfo(envString, defaultValue string) string {
	godotenv.Load(".env")
	envValue := os.Getenv(envString)
	if envValue != "" {
		return envValue
	}

	return defaultValue

}

const (
	RTF_START = `{\rtf1\ansi\ansicpg1252\uc1\deff0{\fonttbl
{\f0\fnil\fcharset0\fprq2 Arial;}
{\f1\fswiss\fcharset0\fprq2 Segoe UI;}
{\f2\froman\fcharset2\fprq2 Symbol;}}
{\colortbl;\red0\green0\blue0;\red255\green255\blue255;}
{\stylesheet{\s0\itap0\nowidctlpar\f0\fs24 [Normal];}{\*\cs10\additive Default Paragraph Font;}}
{\*\generator TX_RTF32 15.1.531.502;}
\deftab1134\paperw12240\paperh15840\margl1440\margt1440\margr1440\margb1440\widowctrl\formshade\sectd
\headery720\footery720\pgwsxn12240\pghsxn15840\marglsxn1440\margtsxn1440\margrsxn1440\margbsxn1440\pard\itap0\nowidctlpar\plain\f1\fs16`
	RTF_END = `\par }`
)
