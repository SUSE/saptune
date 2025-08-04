# create markdown table from coverage output
ls -l *.out *html
cov_tab="coverage_table.md"

echo "| File | Coverage | Status |" > $cov_tab
echo "| --- | --- | --- |" >> $cov_tab

cat cov.html | grep '<option value="file' | sed -E 's/.*>(.*) \((.*)%\)<.*/\2 \1/' | sort -rn |\
while read -r cov file; do
    status=":cloud:"
    if (( $(echo "$cov <= 50" | bc -l) )) ; then
        status=":fire:"
    elif (( $(echo "$cov > 80" | bc -l) )); then
        status=":sunny:"
    fi
    echo "| $file | $cov | $status |" >> $cov_tab
done

