package test

func Tx1() {
	SendRefName("A", "B", 5)
}
func Tx2() {
	SendRefName("B", "D", 1)
	SendRefName("A", "C", 1)
}
func Tx3() {
	SendRefName("B", "D", 1)
	SendRefName("D", "A", 1)
}
