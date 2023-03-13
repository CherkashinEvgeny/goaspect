package test

import io1 "io"

type Face interface {
	Reader() (hehe io1.Reader)
	Reader1() (hehe io1.Reader)
	Reader2() (hehe io1.Reader)
	Reader3() (hehe io1.Reader)
}
