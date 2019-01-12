cd server && ./bstart.sh && cd -
cd qps_client && ./bstart.sh 172.17.32.174 && cd -
cd qps_client && ./bstart.sh 172.17.32.175 && cd -
cd delay_client && ./bstart.sh 172.17.32.176 && cd -
#cd real_client && ./bstart.sh 172.16.186.216 && cd -
#cd client && ./bstart.sh 172.16.186.216 && cd -
#cd client && ./bstart.sh 172.16.186.218 && cd -
#cd client && ./bstart.sh 172.16.186.219 && cd -
#cd client && ./bstart.sh 172.16.186.220 && cd -
#cd client && ./bstart.sh 172.16.186.221 && cd -
