import logging
import json
import time

import requests

from . import settings


PER_PAGE = 1_000
MAX_RETRIES = 3


def fetch_events(contract_address, fout, max_to_fetch=100_000):
    session = requests.Session()

    params = dict(
        module="account",
        action="tokentx",
        contractaddress=contract_address,
        offset=PER_PAGE,
        sort="desc",
        apikey=settings.ETHERSCAN_API_KEY,
    )

    page = 1
    fetched = 0
    retry_count = 0

    while fetched < max_to_fetch and retry_count <= MAX_RETRIES:
        logging.info("%s/%s", fetched, max_to_fetch)

        params["page"] = page
        res = session.get(settings.ETHERSCAN_API_URL, params=params)

        payload = res.json()
        if payload["status"] != "1":
            time.sleep(2 ** retry_count)
            retry_count += 1
            if retry_count <= MAX_RETRIES:
                logging.warning("failed with status %s: %s, retrying (%s/%s)",
                                payload["status"], payload, retry_count, MAX_RETRIES)
            continue

        retry_count = 0

        if not payload["result"]:
            break

        for event in payload["result"]:
            json.dump(event, fout)
            print(file=fout)

        fetched += len(payload["result"])
        page += 1
