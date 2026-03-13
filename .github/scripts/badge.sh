# pwd - /home/runner/work/saptune/saptune

title="coverage"
color_font="#fff" # default #fff - white
color_title="#555" # default #555 - grey

# #fe7d37 - orange, #e05d44 - red, #4c1 - green, #97ca00 - light green, #9f9f9f - light grey
coverage=$(grep total: cov.out | grep -Eo '[0-9]+\.[0-9]+')
if [ -z "$coverage" ]; then
    color_number="#9f9f9f" # #9f9f9f - light grey
    coverage="unknown"
else
    color_number="#fe7d37" # #fe7d37 - orange
    if (( $(echo "$coverage <= 50" | bc -l) )) ; then
        color_number="#e05d44" # #e05d44 - red
    elif (( $(echo "$coverage > 80" | bc -l) )); then
        color_number="#4c1" # #4c1 - green
    fi
    coverage="${coverage}%"
fi

rm -f badge_cov.svg
cp .github/templates/badge_template.svg badge_cov.svg
sed -i "s/__TITLE__/$title/g" badge_cov.svg
sed -i "s/__NUMBER__/$coverage/g" badge_cov.svg
sed -i "s/__COLOR_TITLE__/$color_title/g" badge_cov.svg
sed -i "s/__COLOR_FONT__/$color_font/g" badge_cov.svg
sed -i "s/__COLOR_NUMBER__/$color_number/g" badge_cov.svg
ls -l badge_cov.svg

