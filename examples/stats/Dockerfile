FROM appropriate/curl
RUN apk update
RUN apk upgrade
RUN apk add --no-cache bash
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh
CMD ["./wait-for-it.sh","stats-agent:3000","--","curl", "http://stats-agent:3000/agent", "-X", "POST", "-H", "Content-Type: application/json;charset=UTF-8", "-H", "token-claim-user: privateserver", "--data", "@/setup.json"]
