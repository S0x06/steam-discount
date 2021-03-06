package monitor


import (
	"strconv"
	"github.com/garyburd/redigo/redis"
	"fmt"
	"encoding/json"
	"time"
)

//当pageChan有值时才进行读操作（即，进行第一次爬虫获取到这个值的时候；启动多个goroutine继续爬取之后的页面数据
func GetContentsByPageChan(pageChan chan int,contentChannel chan MonitorContent){
	maxI,ok := <- pageChan
	if ok {
		for i:=2; i<=maxI ;i++  {
			go GetContent("https://store.steampowered.com/search/?specials=1&page="+strconv.Itoa(i),contentChannel,nil,nil)
		}
		close(pageChan)
	}
}

/**
	当maxContentSize有值时才进行读操作（即，进行第一次爬虫获取到这个值的时候)
	timeout:设置channel读取数据的时候的最大等待时间
 */
func SaveContents(contentChannel chan MonitorContent,maxContentSize chan int,timeout time.Duration){
	i := 0
	var size int
	var c redis.Conn

	if v,f :=<-maxContentSize;f{
		size = v
		//fmt.Println("contentSize",size)
		close(maxContentSize)
		var err error
		c,err = redis.Dial("tcp","127.0.0.1:6379",redis.DialPassword("123456"))
		if err!=nil {
			panic(err)
		}
	}

	to := time.NewTimer(timeout)
	Lable:
	for {
		to.Reset(time.Second)
		select {
		case v := <-contentChannel:
			data , _ := json.Marshal(v)
			_,err := c.Do("HMSET","gameContent",strconv.Itoa(v.Id),string(data))	 //把id作为key，将struct内容转化为json字符串作为value，放到redis里
			if err!=nil {
				fmt.Println("redis error:",err.Error())
				panic(err)
			}
			/*else{
				fmt.Println("reply:",reply)
			}*/
			//fmt.Println(v.Thumbnail,i)
			i++
			if size == i {
				close(contentChannel)
			}
		case <- to.C:
			fmt.Println("读取数据超时","size:",size,"read:",i)
			break Lable
		}
	}

}
