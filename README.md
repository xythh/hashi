 hashi
 ==

A static site generator that has flexible yet simple file structure, is extensibility and supports Japanese pitch notation.

## Features

* Easily write pitch accent with simple syntaxx like べんきょう{0} ,no more hard to search and write  ￣  ＼.
* No forced file hierachy, no complicated file structure to learn, you choose how you want to organize your site.
* Extensible.
* Fast.

## Installation

Download the binaries provided on github or download it using go get.

	# go get github.com/xythh/hashi 


## Supported Markdown

* Basic markdown extended with tables,fenced code blocks,autolinking,strikethrough and a few custom extensions.

* Pitch notation is supported with the syntax WORD{DROP_LOCATION}<br>example: べんきょう{0} will generate the html for べんきょう as a heiban word.

* Headers automatically have an anchor added to them.

* Inlining HTML is supported.

* {#ANCHORNAME} allows you to add a anchor to a table row, which allows you to link to that part of a table.


## Layouts

layouts are placed in the .hashi folder. Layouts are simple HTML with variables that lay out how a markdown document should be build into HTML. The default layout is layout.html but you can set it to another by changing the frontmatter.

``` html
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta content="text/html; charset=UTF-8" http-equiv="Content-Type">
		<meta name="viewport" content="width=device-width,initial-scale=1.0">
		<link href="styles.css" rel="stylesheet" type="text/css">
		<link rel="icon" href="data:,">
	</head>
	<body>
{{content}}
</body>
```
In this example we have a basic layout for our page. At build time, any file that has this as its defined layout, will build following this layout and will replace {{content}} with the files content converted into HTML.

## Frontmatter

Markdown documents can optionally start with a section that defines certain variables that can be used within the Markdown document as well as within the layout that is used by that documnet. Your markdown's frontmatter is written at the start of the document up until  `---` with your content following it.

``` yaml
title: Example
description: This is an example of frontmatter.
layout: example.html

---
```
This sets three variables, which can be accessed in our layout(the layout variable is special, because it changes what layout file we use) or inside of our markdown document. To access a variable simply write {{VARIABLENAME}} and it will be processes accordingly.

``` html
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta content="text/html; charset=UTF-8" http-equiv="Content-Type">
		<meta name="viewport" content="width=device-width,initial-scale=1.0">
		<title>{{title}}</title>
		<link href="styles.css" rel="stylesheet" type="text/css">
		<link rel="icon" href="data:,">
		<meta name="description" content ="{{description}}">
	</head>
	<body>
{{content}}
</body>

```

In this example  at build time {{description}} will be replaced with "This is an example of frontmatter." {{title}} with "example" and {{content}} with your content(everything in the file after your frontmatter with variables processesed).

``` markdown

title: Cool
author: mark
---

# {{title}}

This is a short article written by {{author}} to showcase variables.

```
In this example at build time we get the following.

``` markdown

# Cool

This is a short article written by mark to showcaase variables.

```
This is then converted into html.

## Default variables
| variable    | Default value                                                                                        |
|-------------|-------------------------------------------------------------------------------------------------     |
| title       | The filename in all caps.                                                                            |
| description | this is empty by default                                                                             |
| file        | the name of the file with the file extension                                                         |
| url         | for markdown files it defaults to FILENAME.html for other files it's equal    to file                |
| output      | .pub/filename.html                                                                                   |
| layout      | layout.html                                                                                          |
| content     | the contents of your markdown file converted to HTML, NOT recommended to override this variable      |

> **Warning**
 Your layout variable must be a valid layout file, if it is not your file will not be build.
 
 
## Commands
| Command     | arguments          | description                                                                                                                                             |
|-------------|--------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| hashi build | a filename or none | If called with no arguments, it will build all files, if called with a filename it will build only that file and print the built text to the console. |
| hashi watch | no arguments       | Builds all files,keeps running and builds files as they are modified.                                                                                    |
| hashi var   | a filename         | Prints the variables for the file.                                                                                                                      |
 
 ## Publication folder
 
 Running `hashi build` or `hashi watch` for the first time will created a `.pub` folder which has the same file hierachy as the directory that the command is ran from. Markdown files are converted to HTML files, and all other filetypes are directly copied to the `.pub` folder.  `hashi build` and `hashi watch` will build all files ignoring any file or directory that begins with a period.
 
 A example tree is shown below:
 ```
├── cool.md
├── .ignore
├── styles
└── test.md
├── styles
│   └── styles.css
├── .hashi
│   └── layout.html
├── .pub
│   ├── cool.html
│   ├── styles
│   │   └── styles.css
│   ├── styles.css
│   └── test.html

```


# License
Hashi is licensed under the Apache 2.0 License.

