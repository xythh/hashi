# hashi
## Description
A static site generator that converts Markdown into HTML with support for
japanese pitch notation( はし{2} ) and with sane file hierachy.

Turns

![markdown file](markdown.png)

To

![html file](html.png)



# Commands

`hashi build` If ran with no arguments builds all the files in the directory except for files or directories that start with a dot. If ran with an argument it builds only that file.

`hashi watch` bulds files as they are modified.

