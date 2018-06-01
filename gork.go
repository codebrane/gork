// go get github.com/mattn/go-sqlite3

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
)

// Only fields beginning with uppercase are exported for JSON
type Blog struct {
	zEXT_ID        string
	zFEED_LINK     sql.NullString
	zFOLDER_ID     string
	ZTITLE         string `json:"Title"`
	z_ENT          int
	z_OPT          int
	z_PK           int
	z_8FEEDS       int
	z_7FEEDFOLDERS int
	Folder         string
	Feed           string
	Url            string
}

func main() {

	// Sort out the program input
	if (len(os.Args)) < 3 {
		fmt.Println("Usage:")
		fmt.Println("gork PATH_TO_READKIT_DATABASE PATH_TO_OUTPUT_FILE")
		fmt.Println("e.g.")
		fmt.Println("gork ReadKit.storedata blogs.json")
		os.Exit(1)
	}
	readkitDatabase := os.Args[1]
	blogsFile := os.Args[2]

	// Keep a running list of all blogs
	blogs := make(map[int]Blog)

	// Open the ReadKit database
	db, err := sql.Open("sqlite3", readkitDatabase)
	checkErr(err)

	// Load up all the blogs from the database into our blogs map
	rows, err := db.Query("SELECT ZEXT_ID,ZFEED_LINK,ZFOLDER_ID,ZTITLE,Z_ENT,Z_OPT,Z_PK FROM ZFOLDER where ZEXT_ID is not null and ZFOLDER_ID is not null and ZTITLE is not null and Z_ENT is not null and Z_OPT is not null and Z_PK is not null")
	checkErr(err)
	for rows.Next() {
		blog := new(Blog)
		err = rows.Scan(&blog.zEXT_ID, &blog.zFEED_LINK, &blog.zFOLDER_ID, &blog.ZTITLE, &blog.z_ENT, &blog.z_OPT, &blog.z_PK)
		blog.Url = blog.zFEED_LINK.String
		blog.Feed = strings.Replace(blog.zEXT_ID, "feed/", "", 1)
		checkErr(err)
		blogs[blog.z_PK] = *blog
	}
	rows.Close()

	// Hook up each blog with its folder
	for z_PK, blog := range blogs {
		stmt, err := db.Prepare("SELECT Z_8FEEDS,Z_7FEEDFOLDERS from Z_8FEEDFOLDERS WHERE Z_8FEEDS = ?")
		checkErr(err)
		rows, err := stmt.Query(z_PK)
		for rows.Next() {
			err = rows.Scan(&blog.z_8FEEDS, &blog.z_7FEEDFOLDERS)
			blog.Folder = blogs[blog.z_7FEEDFOLDERS].ZTITLE
			blogs[blog.z_PK] = blog
		}
	}
	rows.Close()

	db.Close()

	// Delete any existing blogs output file
	if _, err := os.Stat(blogsFile); err == nil {
		err = os.Remove(blogsFile)
	}
	f, err := os.OpenFile(blogsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	checkErr(err)
	defer f.Close()

	totalBlogs := len(blogs)
	currentBlog := 0

	// Create the blogs output file
	_, err = f.WriteString("[\n")
	checkErr(err)
	for _, blog := range blogs {
		currentBlog++
		if blog.z_ENT == 8 {
			t, err := json.MarshalIndent(blog, "", "  ")
			_, err = f.WriteString(string(t))
			checkErr(err)
			if currentBlog < totalBlogs {
				_, err = f.WriteString(",")
				checkErr(err)
			}
			_, err = f.WriteString("\n")
			checkErr(err)
		}
	}
	_, err = f.WriteString("]\n")
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
