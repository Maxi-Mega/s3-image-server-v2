# Expr Functions

## _call

_Call another expr_

`_call(exprName string) (any, error)`

## _exist

_Check whether a file exists in cache_

`_exist(fileSelector string) (bool, error)`

## _jq

`_jq(fileSelector string, filter string) (any, error)`

## _loadJSON

`_loadJSON(fileSelector string) (any, error)`

## _merge

`_merge(o1 map[string]any, o2 map[string]any) (map[string]any, error)`

## _s3Key

`_s3Key(fileSelector string) (string, error)`

## _xpath

`_xpath(fileSelector string, xpath string) (any, error)`

