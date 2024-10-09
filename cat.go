package main

func main1() {
	x, y, z := function(1, 10, 100)
	println(x, y, z)
}

func function(a, b, c int64) (int64, int64, int64) {
	defer func() {
		a = a + 2
	}()

	a++
	b++
	c++
	return a, b, c
}

func demo1() int {
	ret := -1
	defer func() {
		ret = 1
	}()
	return ret
}

func demo2() (ret int) {
	defer func() {
		ret = 1
	}()
	return ret
}
