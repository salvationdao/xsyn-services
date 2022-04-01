#!/usr/bin/python3
# Standard Lib
from datetime import datetime
import tarfile
import sys
import os
import getopt
import json
import logging
import shutil
import re
import gzip
import pathlib
import subprocess
try:
    import pip
except ImportError as e:
    print('''
    Install python3-pip
    apt-get install python3-pip
    ''')
    exit(1)

# Pip Installs
try:
    import requests
    from tqdm import tqdm
except ImportError as e:
    pip.main(['install', e.name])

    print("\n Dependency installed rerun script")
    exit(1)


REPO = 'ninja-syndicate/passport-server'
BASE_URL = "https://api.github.com/repos/{repo}".format(repo=REPO)
TOKEN = os.environ.get("GITHUB_PAT", "")
CLIENT = "ninja_syndicate"
BASE_DIR = "/usr/share/{client}".format(client=CLIENT)
PACKAGE = "passport-api"


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
    # Load Env
    load_package_env(
        "{package}_online/init/{package}.env".format(package=PACKAGE))

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

    # Extract asset
    if not question("Extract {} or exit?".format(rel_path), 'extract', 'exit'):
        log.info("exiting")
        exit(0)

    new_ver_dir = extract(rel_path)

    copy_env(new_ver_dir)
    dbdump()


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
        if os.path.exists(fname):
            if not question("{} exists, overwrite?".format(fname)):
                log.info("Skipping Download")
                return fname

        log.info("Downloading: %s", fname)
        log.debug("code: %s", resp.status_code)
        log.debug("headers: %s", resp.headers)
        progress_bar = tqdm(total=file_size, unit='iB', unit_scale=True)
        with open(file_name, 'wb') as f:
            for chunk in resp.iter_content(chunk_size=8192):
                f.write(chunk)
                progress_bar.update(len(chunk))
        progress_bar.close()

    log.info("Downloaded: %s", os.path.abspath(file_name))
    return file_name


def extract(file_name: str):
    if not question("Extract {} or exit?".format(file_name), negative='exit'):
        log.info("exiting")
        exit(0)

    log.info("Extract: {}".format(file_name))
    dest = file_name.rstrip(".tar.gz")
    if os.path.exists(dest):
        if not question("Destination exists, overwrite?"):
            log.info("Skipping extraction")
            return dest

    if file_name.endswith("tar.gz"):
        tar = tarfile.open(file_name, "r:gz")
        tar.extractall()
        tar.close
        return dest


def load_package_env(env_file):
    try:
        with open(env_file) as f:
            for line in f:
                if line.startswith('#') or not line.strip():
                    continue
                if 'export' in line:
                    # Remove leading `export `
                    line = line.removesuffix("export ")

                key, value = line.strip().split('=', 1)
                os.environ[key] = value.strip('"')  # Load to local environ
    except FileNotFoundError as e:
        log.exception("file not found: %s", e.filename)
        exit(1)

    log.info("loaded env vars from %s", env_file)


def copy_env(target: str):
    src = "{package}_online/init/{package}.env".format(package=PACKAGE)
    dest = "{target}/init/{package}.env".format(target=target, package=PACKAGE)
    log.debug("src: %s", src)
    log.debug("target: %s", target)
    log.debug("dest: %s", dest)
    try:
        shutil.copyfile(src, dest)
    except FileNotFoundError as e:
        log.exception("file not found: %s", e.filename)
        exit(1)
    log.info("Coppied " + src + " to " + dest)


def dbdump():
    if question("Skip database dump"):
        log.info("Skipping database dump")
        return False

    log.info("Starting database dump")

    dump_dir = "{base_dir}/{package}_online/db_copy".format(
        base_dir=BASE_DIR, package=PACKAGE)
    pathlib.Path(dump_dir).mkdir(parents=True, exist_ok=True)

    now = datetime.now()
    dump_file = "{dump_dir}/{package}_{now}.sql.gz".format(
        dump_dir=dump_dir,
        package=PACKAGE,
        now=now.strftime("%Y%m%d%H%M%S"))

    command = 'pg_dump --dbname="{dbname}" --host="{host}" --port="{port}" --username="postgres" '.format(
        dbname=os.environ.get("PASSPORT_DATABASE_NAME"),
        host=os.environ.get("PASSPORT_DATABASE_HOST"),
        port=os.environ.get("PASSPORT_DATABASE_PORT"))

    try:
        with gzip.open(dump_file, 'wb') as f:
            popen = subprocess.Popen(
                command, stdout=subprocess.PIPE, shell=True, universal_newlines=True)

            for stdout_line in iter(popen.stdout.readline, ''):
                f.write(stdout_line.encode('utf-8'))

            popen.stdout.close()
            popen.wait()
    except FileNotFoundError as e:
        log.exception("file not found: %s", e.filename)
        exit(1)

    log.info("Dumped database " +
             os.environ.get("PASSPORT_DATABASE_NAME") + " into " + dump_file)

    if not os.path.exists(dump_file):
        log.error("Dump file doesn't exist")
        exit(1)
    if not os.path.getsize(dump_file) > 5e4:
        log.error("Dump file smaller that expected")
        exit(1)


def question(question, positive='y', negative='n'):
    question = question + \
        ' ({positive}/{negative}): '.format(positive=positive, negative=negative)
    while "the answer is invalid":
        reply = str(input(question)).lower().strip()
        log.debug("reply %s", reply)
        if reply == positive:
            return True
        if reply == negative:
            return False


if __name__ == "__main__":
    main(sys.argv[1:])
