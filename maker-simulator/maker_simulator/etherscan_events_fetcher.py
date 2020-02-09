import logging
import json
import time

import requests

from . import settings


PER_PAGE = 1_000
MAX_RETRIES = 3


def fetch_events(contract_address, fout, max_to_fetch=100_000):
    page = 1
    params = dict(
        module="account",
        action="tokentx",
        contractaddress=contract_address,
        page=page,
        offset=PER_PAGE,
        sort="desc",
        apikey=settings.ETHERSCAN_API_KEY,
    )
    session = requests.Session()
    fetched = 0
    retry_count = 0
    while fetched < max_to_fetch and retry_count <= MAX_RETRIES:
        logging.info("%s/%s", fetched, max_to_fetch)
        res = session.get(settings.ETHERSCAN_API_URL, params=params)
        payload = res.json()
        if payload["status"] != "1":
            time.sleep(2 ** retry_count)
            retry_count += 1
            if retry_count <= MAX_RETRIES:
                logging.warning("failed with status %s, retrying (%s/%s)",
                                payload["status"], retry_count, MAX_RETRIES)
            continue

        retry_count = 0
        for event in payload["result"]:
            json.dump(event, fout)
            print(file=fout)
        fetched += len(payload["result"])
        page += 1
