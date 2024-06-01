import sys
raw = sys.argv[1]
arr = raw.split(",")
arr = [x.strip() for x in arr]
print(arr[0])
