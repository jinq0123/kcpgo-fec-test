package main

type Mode struct {
	name string

	nodelay  int
	interval int
	resend   int
	nc       int

	fec bool
}
