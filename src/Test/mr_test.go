/*
 * @Author: your name
 * @Date: 2021-07-04 23:50:45
 * @LastEditTime: 2021-07-05 08:45:41
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /6.824_Assignment/src/Test/mr_test.go
 */
package Test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"testing"
)

func TestOSPackage(t *testing.T) {

	type test struct {
		Name string `json:"name"`
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Printf("os Getwd Err")
		panic(err)
	}
	fmt.Println("pwd is - ", pwd)

	TestFile := pwd + "/" + "test.json"
	var file *os.File

	file, err = os.OpenFile(TestFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("file create err")
	}
	defer file.Close()

	tt := &test{"1"}
	tt2 := &test{"2"}
	tt3 := &test{"3"}
	encode := json.NewEncoder(file)
	encode.Encode(tt)
	encode.Encode(tt2)
	encode.Encode(tt2)
	encode.Encode(tt3)

	// fmt.Fprintf(file, "%v %v\n", "1", "1")
	// file.WriteString("ttt")
	// file.Write([]byte("ttttttttt"))
	// file.Sync()
	// file.Close()
}

// This Test proft
func TestForFileLock(t *testing.T) {
	// pwd, err := os.Getwd()
	// if err != nil {
	// 	log.Printf("os Getwd Err")
	// 	panic(err)
	// }
	// fmt.Println("pwd is - ", pwd)

	// TestFile := pwd + "/" + "test.txt"
	// var file *os.File
	// if checkFileIsExist(TestFile) {
	// 	file, err = os.OpenFile(TestFile, os.O_APPEND, 0666)
	// 	if err != nil {
	// 		log.Printf("file create err")
	// 	}
	// 	// defer file.Close()
	// } else {
	// 	file, err = os.Create(TestFile)
	// 	if err != nil {
	// 		log.Printf("file create err")
	// 	}
	// 	// defer file.Close()
	// }
	// defer file.Close()

	// add lock to file
	// // 非阻塞模式下，加共享锁
	// if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
	// 	log.Println("add share lock in no block failed", err)
	// }
	// // 这里进行业务逻辑
	// // TODO

	// // 解锁
	// if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
	// 	log.Println("unlock share lock failed", err)
	// }

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			pwd, err := os.Getwd()
			if err != nil {
				log.Printf("os Getwd Err")
				panic(err)
			}
			fmt.Println("pwd is - ", pwd)

			TestFile := pwd + "/" + "test.txt"
			var file *os.File

			file, err = os.OpenFile(TestFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				log.Printf("OpenFile Error")
			}
			defer file.Close()

			// if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
			// 	log.Println("add share lock in no block failed", err)
			// 	return
			// }

			writeStr := "Test" + strconv.Itoa(index) + "\n"
			file.Write([]byte(writeStr))
			// file.Sync()
			// // 解锁
			// if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
			// 	log.Println("unlock share lock failed", err)
			// 	return
			// }

		}(i)
	}

	wg.Wait()

}

func TestWriteAppend(t *testing.T) {

	filePath := "/Users/bytedance/6.824_Assignment/src/Test/test.txt"
	// 追加打开
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println("open file error")
	}
	defer file.Close()

	file.Write([]byte("TestWriteAppend!"))
	file.Sync()
}

/*
 Utils contains some useful methonds.
*/

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

// sync map
func TestForSync(t *testing.T) {
	var testmap sync.Map
	var result string

	for i := 0; i < 10; i++ {
		testmap.Store(i, "key"+strconv.Itoa(i))
	}

	// TODO: 为什么这个是闭包函数？
	testmap.Range(func(key, value interface{}) bool {
		if key == 4 {
			result = value.(string)
			return false
		}
		return true
	})

	fmt.Println(result)
}

func TestLock(t *testing.T) {
	mutex := sync.Mutex{}
	cond := sync.NewCond(&mutex)

	cond.Wait()
}
