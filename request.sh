#!/bin/bash

curl --url 'https://solr.sscdn.co/cc/c7/?q=um+dia+um+adeus&limit=30&callback=suggest_callback' \
  -H 'accept: */*' \
  -H 'accept-language: en-US,en;q=0.9,pt;q=0.8' \
  -H 'origin: https://www.cifraclub.com.br' \
  -H 'priority: u=1, i' \
  -H 'referer: https://www.cifraclub.com.br/' \
  -H 'sec-ch-ua: "Not?A_Brand";v="24", "Chromium";v="152"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "Linux"' \
  -H 'sec-fetch-dest: empty' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-site: cross-site' \
  -H 'user-agent: Mozilla/5.0 (X11; Ubuntu; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36'
