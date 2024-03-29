package arrayOpr

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/stellentus/cartpoles/lib/rlglue"
)


func FloatAryToString(a []float64, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

func Search2D(target []float64, list [][]float64) int {
	for idx, l := range list {
		if EqualArrys(l, target) {
			return idx
		}
	}
	return -1
}

func SearchInt(a int, list []int) int {
	for idx, b := range list {
		if b == a {
			return idx
		}
	}
	return -1
}

func EqualArrys(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func AllEqualInt(arr []int) bool {
	var equal bool
	equal = true
	for i := 1; i < len(arr); i++ {
		if arr[i] != arr[0] {
			equal = false
		}
	}
	return equal
}

func Concatenate(arr1, arr2 [][]float64) [][]float64 {
	if len(arr1) != len(arr2) {
		fmt.Printf("Concatenate: Arrays should have the same length. Length arr1 = %d, length arr2 = %d\n", len(arr1), len(arr2))
		os.Exit(1)
	}
	new := make([][]float64, len(arr1))
	dim1 := len(arr1[0])
	dim2 := len(arr2[0])
	for i := 0; i < len(new); i++ {
		new[i] = make([]float64, dim1+dim2)
		copy(new[i][:dim1], arr1[i])
		copy(new[i][dim1:], arr2[i])
	}
	return new
}

func A32Col(array [][]float32, col int) []float32 {
	var new []float32
	for i := 0; i < len(array); i++ {
		// new[i] = array[i][col]
		new = append(new, array[i][col])
	}
	return new
}

/*
Take block from 2-d array
*/
func Index2d(array [][]float64, rowStart int, rowEnd int, colStart int, colEnd int) [][]float64 {
	var new [][]float64
	for i := rowStart; i < rowEnd; i++ {
		var temp []float64
		new = append(new, temp)
		for j := colStart; j < colEnd; j++ {
			// new[i-rowStart][j-colStart] = array[i][j]
			new[i] = append(new[i], array[i][j])
		}
	}
	return new
}

/*
sample from 1-d array
*/
func SampleByIdx1dInt(array []int, rowIdx []int) []int {
	res := make([]int, len(rowIdx))
	for i := 0; i < len(rowIdx); i++ {
		res[i] = array[rowIdx[i]]
	}
	return res
}
/*
sample from 2-d array
*/
func SampleByIdx2d(array [][]float64, rowIdx []int) [][]float64 {
	var res [][]float64
	for i := 0; i < len(rowIdx); i++ {
		temp := make([]float64, len(array[rowIdx[i]]))
		copy(temp, array[rowIdx[i]])
		res = append(res, temp)
	}
	return res
}
/*
sample from 3-d array
*/
func SampleByIdx3d(array [][][]float64, rowIdx []int) [][][]float64 {
	var res [][][]float64
	for i := 0; i < len(rowIdx); i++ {
		temp := make([][]float64, len(array[rowIdx[i]]))
		copy(temp, array[rowIdx[i]])
		res = append(res, temp)
	}
	return res
}

/*
Index in each row of 2-d array
*/
func RowIndexFloat(array [][]float64, idx []int) []float64 {
	// var new []float64
	// for i := 0; i < len(array); i++ {
	// 	new = append(new, array[i][idx[i]])
	// }
	new := make([]float64, len(array))
	for i := 0; i < len(array); i++ {
		new[i] = array[i][idx[i]]
	}
	return new
}

/*
Max in each row of 2-d array
*/
func RowIndexMax(array [][]float64) ([]float64, []int) {
	new := make([]float64, len(array))
	arg := make([]int, len(array))
	for i := 0; i < len(array); i++ {
		max, idx := ArrayMax(array[i])
		new[i] = max
		arg[i] = idx
	}
	return new, arg
}

/*
Max of 1-d array
*/
func ArrayMax(array []float64) (float64, int) {
	max := math.Inf(-1)
	var idx int
	for j := 0; j < len(array); j++ {
		if array[j] > max {
			max = array[j]
			idx = j
		}
	}
	return max, idx
}

/*
Min of 1-d array
*/
func ArrayMin(array []float64) (float64, int) {
	min := math.Inf(1)
	var idx int
	for j := 0; j < len(array); j++ {
		if array[j] < min {
			min = array[j]
			idx = j
		}
	}
	return min, idx
}

/*
Max of a column in a 2-d array
*/
func ColumnMax(array [][]float64, col int) (float64, int) {
	max := math.Inf(-1)
	var idx int
	for j := 0; j < len(array); j++ {
		if array[j][col] > max {
			max = array[j][col]
			idx = j
		}
	}
	return max, idx
}

/*
Min of a column in a 2-d array
*/
func ColumnMin(array [][]float64, col int) (float64, int) {
	min := math.Inf(1)
	var idx int
	for j := 0; j < len(array); j++ {
		if array[j][col] < min {
			min = array[j][col]
			idx = j
		}
	}
	return min, idx
}

func StateTo32(state rlglue.State) []float32 {
	var a32 []float32
	for i := 0; i < len(state); i++ {
		a32 = append(a32, float32(state[i]))
	}
	return a32
}

func A64To32_2d(array [][]float64) [][]float32 {
	var a32 [][]float32
	for i := 0; i < len(array); i++ {
		var temp []float32
		a32 = append(a32, temp)
		for j := 0; j < len(array[i]); j++ {
			a32[i] = append(a32[i], float32(array[i][j]))
		}
	}
	return a32
}

func A64To32(array []float64) []float32 {
	var a32 []float32
	for i, f64 := range array {
		a32[i] = float32(f64)
	}
	return a32
}

func A32To64(array []float32) []float64 {
	var a64 []float64
	for i, f32 := range array {
		a64[i] = float64(f32)
	}
	return a64
}

func A32ToInt(array []float32) []int {
	var aInt []int
	for i, f32 := range array {
		aInt[i] = int(f32)
	}
	return aInt
}

func A64ToInt2D(array [][]float64) [][]int {
	var aInt [][]int
	for i := 0; i < len(array); i++ {
		aInt = append(aInt, A64ToInt(array[i]))
	}
	return aInt
}
func A64ToInt(array []float64) []int {
	var aInt []int
	for _, f64 := range array {
		aInt = append(aInt, int(f64))
	}
	return aInt
}

func IntToA642D(array [][]int) [][]float64 {
	aFloat := make([][]float64, len(array))
	for i := 0; i < len(array); i++ {
		aFloat[i] = make([]float64, len(array[i]))
		copy(aFloat[i], IntToA64(array[i]))
	}
	return aFloat
}
func IntToA64(array []int) []float64 {
	var aFloat []float64
	for _, i := range array {
		aFloat = append(aFloat, float64(i))
	}
	return aFloat
}

func A64ArrayMulti2D(a float64, arr [][]float64) [][]float64 {
	res := make([][]float64, len(arr))
	for i := 0; i < len(arr); i++ {
		res[i] = make([]float64, len(arr[i]))
		copy(res[i], A64ArrayMulti(a, arr[i]))
	}
	return res
}
func A64ArrayMulti(a float64, arr []float64) []float64 {
	res := make([]float64, len(arr))
	for i := 0; i < len(arr); i++ {
		res[i] = a * arr[i]
	}
	return res
}

func BitwiseAdd2D(a [][]float64, b [][]float64) [][]float64 {
	if len(a) != len(b) {
		errors.New("Arrays should have same length")
	}
	res := make([][]float64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = make([]float64, len(a[i]))
		copy(res[i], BitwiseAdd(a[i], b[i]))
	}
	return res
}
func BitwiseAdd(a []float64, b []float64) []float64 {
	if len(a) != len(b) {
		errors.New("Arrays should have same length")
	}
	res := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = a[i] + b[i]
	}
	return res
}

func BitwiseMulti2D(a [][]float64, b [][]float64) [][]float64 {
	if len(a) != len(b) {
		errors.New("Arrays should have same length")
	}
	res := make([][]float64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = make([]float64, len(a[i]))
		copy(res[i], BitwiseMulti(a[i], b[i]))
	}
	return res
}
func BitwiseMulti(a []float64, b []float64) []float64 {
	res := make([]float64, len(a))
	if len(a) != len(b) {
		errors.New("Arrays should have same length")
	}
	for i := 0; i < len(a); i++ {
		res[i] = a[i] * b[i]
	}
	return res
}

func BitwiseMinus2D(a [][]float64, b [][]float64) [][]float64 {
	if len(a) != len(b) {
		errors.New("Arrays should have same length")
	}
	res := make([][]float64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = make([]float64, len(a[i]))
		copy(res[i], BitwiseMinus(a[i], b[i]))
	}
	return res
}
func BitwiseMinus(a []float64, b []float64) []float64 {
	if len(a) != len(b) {
		errors.New("Arrays should have same length")
	}
	res := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = a[i] - b[i]
	}
	return res
}

func BitwiseDivide(a []float64, b []float64) []float64 {
	if len(a) != len(b) {
		errors.New("Arrays should have same length")
	}
	res := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = a[i] / b[i]
	}
	return res
}

func BitwisePower2D(array [][]float64, p float64) [][]float64 {
	res := make([][]float64, len(array))
	for i := 0; i < len(array); i++ {
		res[i] = make([]float64, len(array[i]))
		copy(res[i], BitwisePower(array[i], p))
	}
	return res
}
func BitwisePower(array []float64, p float64) []float64 {
	res := make([]float64, len(array))
	for i := 0; i < len(array); i++ {
		res[i] = math.Pow(array[i], p)
	}
	return res
}


func ReSize1DInt(inputData []int, rows, columns int) [][]int {
	if rows*columns != len(inputData) {
		fmt.Println("ReSize1DInt: Wrong Size")
		os.Exit(1)
	}
	res := make([][]int, rows)
	for i := 0; i < rows; i++ {
		res[i] = make([]int, columns)
		copy(res[i], inputData[rows*columns:rows*(columns+1)])
	}
	return res
}

func ReSize1DA64(inputData []float64, rows, columns int) [][]float64 {
	if rows*columns != len(inputData) {
		fmt.Println("ReSize1DA64: Wrong Size")
		os.Exit(1)
	}
	res := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		res[i] = make([]float64, columns)
		copy(res[i], inputData[rows*columns:rows*(columns+1)])
	}
	return res
}

func Flatten2DFloat(inputData [][]float64) []float64 {
	var flatten []float64
	for i := 0; i < len(inputData); i++ {
		for j := 0; j < len(inputData[0]); j++ {
			flatten = append(flatten, inputData[i][j])
		}
	}
	return flatten
}

func Flatten2DInt(inputData [][]int) []int {
	var flatten []int
	for i := 0; i < len(inputData); i++ {
		for j := 0; j < len(inputData[0]); j++ {
			flatten = append(flatten, inputData[i][j])
		}
	}
	return flatten
}

func Sum(inputData []float64) float64 {
	sum := 0.0
	for i := 0; i < len(inputData); i++ {
		sum = sum + inputData[i]
	}
	return sum
}
func Average(inputData []float64) float64 {
	sum := Sum(inputData)
	sum = sum / float64(len(inputData))
	return sum
}

func SumOnAxis2D(inputData [][]float64, axis int) []float64 {
	temp := make([]float64, len(inputData))
	for i:=0; i<len(inputData); i++ {
		temp[i] = Sum(inputData[i])
	}
	if axis == 1 {
		return temp
	} else if axis==0 {
		sum := make([]float64, 1)
		sum[0] = Sum(temp)
		return sum
	} else {
		return nil
	}
}

func Absolute2D(inputData [][]float64) [][]float64 {
	res := make([][]float64, len(inputData))
	for i:=0; i<len(inputData); i++ {
		res[i] = make([]float64, len(inputData[i]))
		res[i] = Absolute(inputData[i])
	}
	return res
}
func Absolute(inputData []float64) []float64 {
	res := make([]float64, len(inputData))
	for i:=0; i<len(inputData); i++ {
		res[i] = math.Abs(inputData[i])
	}
	return res
}

func OneHotSet(idx float64, bin int) []float64 {
	res := make([]float64, bin)
	res[int(idx)] = 1
	return res
}
func OneHotSet2D(idx []float64, bin int) [][]float64 {
	res := make([][]float64, len(idx))
	for i:=0; i<len(idx); i++ {
		res[i] = OneHotSet(idx[i], bin)
	}
	return res
}