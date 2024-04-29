#!/bin/bash
arr=(maple-culvert-tracker-bot maple-culvert-tracker-web maple-culvert-tracker-chartmaker maple-culvert-tracker-periodicredis)
# theres also the reminder thats missing here, but it shouldnt matter much
for i in "${arr[@]}"; do
    echo "Importing $i"
    docker image load <"$HOME/$i.dockerimage"
done
