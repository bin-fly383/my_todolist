package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

type Todo struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status bool   `json:"status"`
}

func initPostgreSQL() (err error) {
	dsn := "host=localhost user=postgres password=pg290430 dbname=bubble port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return checkDB(DB)
}

func checkDB(gormDB *gorm.DB) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func simpleClose(gormDB *gorm.DB) {
	sql, err := gormDB.DB()
	if err != nil {
		log.Printf("failed to close database:%v", err)
	} else {
		sql.Close()
	}
}

func main() {
	// 创建数据库 在postgreSQL中 CREATE DATABASE bubble;
	// 连接数据库
	err := initPostgreSQL()
	if err != nil {
		panic(err)
	}
	defer simpleClose(DB)
	r := gin.Default()

	// 模型绑定
	DB.AutoMigrate(&Todo{})

	// 路由组
	v1Group := r.Group("v1")
	{
		// 待办事项

		// 添加
		v1Group.POST("/todo", func(c *gin.Context) {
			// 前端页面填写待办事项 点击提交 会发请求到这里
			// 从请求中把数据拿出来
			var todo Todo
			c.BindJSON(&todo)

			// 返回响应
			if err = DB.Create(&todo).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, todo)
			}
		})
		// 查看所有的待办事项
		v1Group.GET("/todo", func(c *gin.Context) {
			// 从数据库读取数据
			var todolist []Todo // 声明一个结构体切片变量，用于存储从数据库找到的所有数据
			if err := DB.Find(&todolist).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, todolist)
			}

		})
		// 查看某一个待办事项
		v1Group.GET("/todo/:id", func(c *gin.Context) {
			// 获取路径参数
			id := c.Param("id")
			// 从数据库读取数据
			var todo Todo
			if err := DB.Where("ID=?", id).First(&todo).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, todo)
			}
		})
		// 修改
		v1Group.PUT("todo/:id", func(c *gin.Context) {
			// 获取路径参数
			id := c.Param("id")
			var todo Todo
			if err := DB.Where("ID=?", id).First(&todo).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
				return
			}
			c.BindJSON(&todo)
			if err := DB.Save(&todo).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{id: "updated"})
			}
		})
		// 删除
		v1Group.DELETE("todo/:id", func(c *gin.Context) {
			// 获取路径参数
			id := c.Param("id")

			if err := DB.Where("ID=?", id).Delete(&Todo{}).Error; err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusOK, gin.H{id: "deleted"})
			}
		})
	}

	r.Run(":9090") //前端的proxy配置为9090端口
}
