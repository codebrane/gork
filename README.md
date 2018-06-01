# gork
Go ReadKit.

Export blogs and folders from ReadKit.

## Why?
After putting a new hard disk in my ageing MacBook Pro and reinstalling [ReadKit](https://readkitapp.com/) I remembered I'd deleted the Google account it was using to sync feeds. Strangely enough, it was still running fine on the even older Mac and that's when I realised there was no way to export my list of blogs and their folders from ReadKit. Having deleted my Google account there was also no way to login to "do something".

According to the [ReadKit FAQ](https://readkitapp.com/help/) there is a way to reset by deleting the backend store. That looked intriguing so I loaded it up in vi and lo and behold! it turned out to be a SQL Format 3 database.

So I opened it in a sqlite viewer, exported it to JSON and came to the conclusion:

from the ZFOLDER table we have a blog entry and a folder entry:
<pre>
{
	"ZACCOUNT": "2",
	"ZBADGEINDEX": "",
	"ZDATE_UPDATED": "",
	"ZEXT_ID": "feed/http://feeds.feedburner.com/ServiceArchitecture",
	"ZFEED_LINK": "http://service-architecture.blogspot.com/",
	"ZFOLDER_ID": "feed/http://feeds.feedburner.com/ServiceArchitecture",
	"ZGROUPINDEX": "",
	"ZIS_EXPANDED": "",
	"ZPOS": "0",
	"ZPREDICATE": "",
	"ZSHOWNOTIFICATION": "",
	"ZSORTINDEX": "",
	"ZTITLE": "Service Architecture - SOA",
	"Z_ENT": "8",
	"Z_OPT": "2",
	"Z_PK": "225"
}

{
	"ZACCOUNT": "2",
	"ZBADGEINDEX": "",
	"ZDATE_UPDATED": "",
	"ZEXT_ID": "user/31b800eb-c70f-4ad0-ab71-cc786a8d9910/category/SOA",
	"ZFEED_LINK": "",
	"ZFOLDER_ID": "feed-folder",
	"ZGROUPINDEX": "",
	"ZIS_EXPANDED": "",
	"ZPOS": "0",
	"ZPREDICATE": "",
	"ZSHOWNOTIFICATION": "",
	"ZSORTINDEX": "",
	"ZTITLE": "SOA",
	"Z_ENT": "7",
	"Z_OPT": "2",
	"Z_PK": "253"
}
</pre>

What the above two entities say is the "Service Architecture - SOA" blog is in the "SOA" folder. How do I know this?

From the Z_PRIMARYKEY table the type of the blog entry can be:
<pre>
"Z_ENT": "8" = Feed
"Z_ENT": "7" = FeedFolder
</pre>

and the Z_8FEEDFOLDERS table holds the links between blogs and their folders:
<pre>
Z_8FEEDS
Z_7FEEDFOLDERS
</pre>

So I knocked up gork to create a JSON file of all my blogs and their folders.

## Usage
I don't recommend running gork on a live ReadKit database file. Copy it somewhere else first. According to the [ReadKit FAQ](https://readkitapp.com/help/) the file is:

<pre>
	~/Library/Containers/com.webinhq.ReadKit/Data/Library/Application Support/ReadKit/ReadKit.storedata
</pre>

so copy it, for example to /tmp/ReadKit.storedata and run the following command. blogs.json will be created in the current working directory:

<pre>
	gork /tmp/ReadKit.storedata blogs.json
</pre>
