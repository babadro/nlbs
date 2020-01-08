package main

type Config struct {
	Port    uint
	DBUser  string
	DBName  string
	DBPass  string
	CleanDB bool
}

var AppConfig Config
