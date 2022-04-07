#!/bin/bash -e

curl 'https://www.random.org/lists/?mode=advanced' \
  -H 'authority: www.random.org' \
  -H 'content-type: application/x-www-form-urlencoded' \
  -H 'accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9' \
  -H 'sec-fetch-site: same-origin' \
  -H 'sec-fetch-mode: navigate' \
  -H 'sec-fetch-user: ?1' \
  -H 'sec-fetch-dest: document' \
  -H 'referer: https://www.random.org/lists/?mode=advanced' \
  --data-raw 'list=Anil%0D%0AMark%0D%0APriya%0D%0AMatt%0D%0AJohn%0D%0AHelen%0D%0AJoe&format=plain&rnd=new' \
  --compressed
