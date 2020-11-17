

x=$(git log --since=midnight --pretty=format:'%an%n'|wc -l)

if [[ $x -gt 0 ]]
then
    echo teste $x
fi
