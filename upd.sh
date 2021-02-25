#!/usr/bin/env bash

update() {

old_version=$1
new_version=$2
okchain_vendor_dir=$3

git fetch
git checkout $new_version
updated_file_list=`git diff $old_version $new_version --stat|awk '{print $1}'`

array=(${updated_file_list// / })

((file_num=${#array[@]}-1))

echo $file_num "files updated"
echo ==========================

for ((i=0; i<${file_num}; i++)) do
    echo "cp ${array[i]} $okchain_vendor_dir/${array[i]}"
    cp ${array[i]} $okchain_vendor_dir/${array[i]}
done

}


update v0.31.9 v0.31.10 /Users/oak/go/src/github.com/ok-chain/okchain/vendor/github.com/tendermint/tendermint