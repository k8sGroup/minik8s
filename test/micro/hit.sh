rm hit
touch hit

for ((i=1;i<=50;i++))
do
    curl '17.17.17.17:8080/ping' >> hit
    echo -e '' >> hit
    sleep 0.1
done

sort hit | uniq -c | sort -rn