#!/usr/bin/python3
# Standard Lib
from fileinput import filename
import tarfile
import sys
import os
import getopt
import json
import logging
import re

# Pip Installs
from tqdm import tqdm
import requests

REPO = 'ninja-syndicate/passport-server'
BASE_URL = "https://api.github.com/repos/{repo}".format(repo=REPO)
TOKEN = os.environ.get("GITHUB_PAT", "")


logging.basicConfig(level=os.environ.get("LOGLEVEL", "INFO"),
                    format="%(levelname)s: %(message)s")
log = logging.getLogger("")


help_msg = '''Usage: ./install_version.py [options...] <version or "latest">
  -h, --help        This help message
  -v, --verbose     Print more logs

Examples:

Get latest
./install_version.py latest

Get version
./install_version.py v1.8.5


Get latest with verbose logging
./install_version.py -v latest
'''


def main(argv):
    if TOKEN == "":
        log.error("Please set GITHUB_PAT environment variable")
        exit(2)

    log.debug("Parsing input")
    inputVersion = ''
    try:
        opts, args = getopt.getopt(argv, "h::v", ["--help", "--verbose"])
    except getopt.GetoptError:
        print(help_msg)
        sys.exit(2)
    if len(args) != 1:
        log.error("There should be one positional argument\n")
        print(help_msg)
        sys.exit(2)

    for opt, arg in opts:
        if opt in ("-h", "--help"):
            print(help_msg)
            sys.exit()
        elif opt in ("-v", "--verbose"):
            log.setLevel(level=logging.DEBUG)

    for arg in args:
        inputVersion = arg

    log.debug("Finished parsing input")

    # Download asset
    asset_meta = download_meta(inputVersion)
    log.debug(asset_meta)
    rel_path = download_asset(asset_meta)
    log.info("Downloaded: %s", os.path.abspath(rel_path))

    # Extract asset
    if not yes_or_no("Extract {} or exit?".format(rel_path)):
        log.info("exiting")
        exit(0)

    extract(rel_path)


def download_meta(version: str):
    headers = {
        'Accept': 'application/vnd.github.v3+json',
        'Authorization': 'token {}'.format(TOKEN),
        'User-Agent': 'python3 http.client'
    }

    release_id = version
    if version != "latest":
        log.info("Getting releases metadata")

        url = "{base}/releases".format(base=BASE_URL)
        res = requests.get(url, headers=headers)
        res.raise_for_status()
        data = res.content
        json_data = json.loads(data.decode("utf-8"))

        release_id = json_data[0]["id"]

    log.info("Getting asset metadata")

    url = "{base}/releases/{release_id}".format(
        base=BASE_URL, release_id=release_id)
    res = requests.get(url, headers=headers)
    res.raise_for_status()

    data = res.content
    json_data = json.loads(data.decode("utf-8"))

    log.debug("asset.id: {}".format(json_data["assets"][0]["id"]))
    log.debug("asset.name: {}".format(json_data["assets"][0]["name"]))
    log.debug("asset.url: {}".format(json_data["assets"][0]["url"]))

    return {
        "id": json_data["assets"][0]["id"],
        "name": json_data["assets"][0]["name"],
        "url": json_data["assets"][0]["url"],
    }


def download_asset(asset_meta: dict):
    log.info("Getting asset: %s", asset_meta["name"])
    url = "{base}/releases/assets/{release_id}".format(
        base=BASE_URL, release_id=asset_meta["id"])
    headers = {
        'Authorization': 'token {}'.format(TOKEN),
        'Accept': 'application/octet-stream',
        'User-Agent': 'python3 http.client'
    }

    file_name = './{}'.format(asset_meta["name"])

    with requests.get(url, headers=headers, stream=True) as resp:
        resp.raise_for_status()
        file_size = int(resp.headers.get("Content-Length"))
        d = resp.headers['content-disposition']
        fname = re.findall("filename=(.+)", d)[0]
        log.info("Downloading: %s", fname)
        log.debug("code: %s", resp.status_code)
        log.debug("headers: %s", resp.headers)
        progress_bar = tqdm(total=file_size, unit='iB', unit_scale=True)
        with open(file_name, 'wb') as f:
            for chunk in resp.iter_content(chunk_size=8192):
                f.write(chunk)
                progress_bar.update(len(chunk))
        progress_bar.close()

    return file_name


def extract(file_name: str):
    log.info("Extract: {}".format(file_name))
    if file_name.endswith("tar.gz"):
        tar = tarfile.open(file_name, "r:gz")
        tar.extractall()
        tar.close

def yes_or_no(question):
    while "the answer is invalid":
        reply = str(input(question+' (y/n): ')).lower().strip()
        if reply[0] == 'y':
            return True
        if reply[0] == 'n':
            return False


if __name__ == "__main__":
    main(sys.argv[1:])
