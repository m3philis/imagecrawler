#! /usr/bin/python

from __future__ import print_function

import re
import os
import os.path
import sys
import http.client
import urllib.request
import time

# regexes
url_regex = re.compile("konachan.(?:com|net)/image/.+?/.+?\.(?:png|jpg)")
name_regex = re.compile("image/.*?/(.*)")

# variable
counter1, counter2 = 0, 15000
tag_filter = None
konachan = None


# little function to calculate the last page of search results
def page_count():
    # open connection to konachan.com
    domain = http.client.HTTPConnection(konachan)

    domain.request("GET", "/post?page=1&tags=" + tags.replace(" ", "+"))
    while True:
        try:
            first_page = domain.getresponse()
            break
        except http.client.BadStatusLine:
            time.sleep(1)
            domain.close()
            domain = http.client.HTTPConnection(konachan)
            domain.request("GET", "/post?page=1&tags=" + tags.replace(" ", "+"))

    # we got our response, now it's time to find that number
    first_page_source = str(first_page.read())
    page_list = first_page_source.split("Next Page")
    number = 0
    for line in page_list:
        if re.search("(?<=\/post\?page\=)\d+", line):
            number = re.search("(?<=\/post\?page\=)\d+", line).group(0)
        else:
            number = 2
    return int(number)

# we don't want to save every picture in one directory.
# so we create a new directory when we donwloaded 15k pics


def directory_size(directory_intern):
    if len(os.listdir(directory_intern)) >= 15000:
        print("Directory " + directory_intern + " full")
        counter1 += 15000
        counter2 += 15000
        directory = "Pics " + str(counter1) + " - " + str(counter2)
        if os.path.isdir(directory):
            print("Directory already exists; skip creation")
        else:
            os.makedirs(directory, 0o755, False)
        os.chdir("..")

# now we start


# user has to set path for pictures
print("Please set download location (full path required): ")
path = sys.stdin.readline()

# set tags, if user want to download specific pictures
print("Set Tags (seperate multiple tags with a whitespace; " +
    "connect tags with more than one word with an underscore): ")
tags = sys.stdin.readline().strip("\n")

# ask if they want to use the safe mode or not
print("Are you wanna use the safe mode of konachan? [yes/no]")
safemode = sys.stdin.readline().strip("\n")

if safemode == "yes":
    konachan = "konachan.net"
else:
    konachan = "konachan.com"

domain = http.client.HTTPConnection(konachan)

# chdir in $path and create directory if it not exists

dir_tags = tags.replace(":", "_")

if not os.path.isdir(path.rstrip()):
    os.makedirs(path.rstrip(), 0o755, True)
os.chdir(path.rstrip())
if safemode == "yes":
    if not os.path.isdir("Safemode Tags " + dir_tags):
        os.makedirs("Safemode Tags " + dir_tags, 0o755, True)
    os.chdir("Safemode Tags " + dir_tags)
else:
    if not os.path.isdir("Tags " + dir_tags):
        os.makedirs("Tags " + dir_tags, 0o755, True)
    os.chdir("Tags " + dir_tags)


# creating directory for pics
directory = "Pics " + str(counter1) + " - " + str(counter2)
if not os.path.isdir(directory):
    os.makedirs(directory, 0o755, True)

# let's start with downloading

for page_number in range(1, page_count()):

    print("Starting download in page " + str(page_number))

    domain.request("GET", "/post?page=" + str(page_number) +
    "&tags=" + tags.replace(" ", "+"))

    while True:
        try:
            index_page = domain.getresponse()
            break
        except http.client.BadStatusLine:
            domain.close()
            domain = http.client.HTTPConnection(konachan)
            domain.request("GET", "/post?page=" + str(page_number) +
             "&tags=" + tags.replace(" ", "+"))
            time.sleep(1)

    # after we got the response from konachan we need the source code
    index_page_source = str(index_page.read())

    # and now we need save every link on this page in a list
    pics_list = index_page_source.split("Post.register")

    directory_size(directory)

    # now we can search every line for the pic link
    for pic in pics_list:
        pic_url = url_regex.search(re.sub("\\\\\\\\", "", pic))

        # if we found the url we download the pic
        # but with whitespaces instead of "%20"
        if pic_url:
            name = name_regex.search(pic_url.group(0)).group(1)
            print("     Downloading pic:  " + name.replace("%20", " ") +
            " in directory: " + directory)

            # a little check if pic already exists
            existance = False
            for dir in os.listdir():
                os.chdir(dir)
                if os.path.isfile(name.replace("%20", " ")):
                    print("     Pic is already on your pc! Skip!")
                    existance = True
                os.chdir("..")

            if not existance:
                os.chdir(directory)
                image = urllib.request.URLopener()
                image.retrieve("http://" +
                    pic_url.group(0), urllib.request.url2pathname(name))
                print("     Download finished")
                os.chdir("..")
