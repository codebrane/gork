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

const FOLDER = 7
const FEED = 8

func main() {

	// Sort out the program input
	if (len(os.Args)) < 3 {
		fmt.Println("Usage:")
		fmt.Println("gork PATH_TO_READKIT_DATABASE PATH_TO_OUTPUT_FILE_WITHOUT_EXTENSION")
		fmt.Println("e.g.")
		fmt.Println("gork ReadKit.storedata blogs")
		fmt.Println("will create blogs.json and blogs.opml")
		os.Exit(1)
	}
	readkitDatabase := os.Args[1]
	blogsFile := os.Args[2]

	// Keep a running list of all blogs...
	blogs := make(map[int]Blog)
	// ...and folders
	folders := make(map[string][]Blog)

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
			folders[blog.Folder] = append(folders[blog.Folder], blog)
		}
	}
	rows.Close()

	db.Close()

	// Delete any existing blogs output files
	if _, err := os.Stat(blogsFile + ".json"); err == nil {
		err = os.Remove(blogsFile + ".json")
	}
	jsonFile, err := os.OpenFile(blogsFile+".json", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	checkErr(err)
	defer jsonFile.Close()

	if _, err := os.Stat(blogsFile + ".opml"); err == nil {
		err = os.Remove(blogsFile + ".opml")
	}
	opmlFile, err := os.OpenFile(blogsFile+".opml", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	checkErr(err)
	defer opmlFile.Close()

	totalBlogs := len(blogs)
	currentBlog := 0

	// Create the blogs OPML file
	_, err = opmlFile.WriteString("<opml version=\"1.0\">\n")
	_, err = opmlFile.WriteString("  <head>\n")
	_, err = opmlFile.WriteString("    <title>OPML</title>\n")
	_, err = opmlFile.WriteString("  </head>\n")
	_, err = opmlFile.WriteString("  <body>\n")

	for folder, blogs := range folders {
		_, err = opmlFile.WriteString("    <outline title=\"" + folder + "\" text=\"" + folder + "\">\n")
		for _, blog := range blogs {
			_, err = opmlFile.WriteString("      <outline text=\"" + blog.ZTITLE + "\" title=\"" + blog.ZTITLE + "\" type=\"rss\" xmlUrl=\"" + blog.Feed + "\"/>\n")
		}
		_, err = opmlFile.WriteString("    </outline>\n")
	}

	_, err = opmlFile.WriteString("  </body>\n")
	_, err = opmlFile.WriteString("</opml>")

	// Create the blogs JSON file
	_, err = jsonFile.WriteString("[\n")
	checkErr(err)
	for _, blog := range blogs {
		currentBlog++
		if blog.z_ENT == FEED {
			t, err := json.MarshalIndent(blog, "", "  ")
			_, err = jsonFile.WriteString(string(t))
			checkErr(err)
			if currentBlog < totalBlogs {
				_, err = jsonFile.WriteString(",")
				checkErr(err)
			}
			_, err = jsonFile.WriteString("\n")
			checkErr(err)
		}
	}
	_, err = jsonFile.WriteString("]\n")
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
