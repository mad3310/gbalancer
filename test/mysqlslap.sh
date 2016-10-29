mysqlslap -h127.0.0.1 -uzbz -pzbz -P3306 --concurrency=200 --iterations=1 --engine=innodb --create-schema='tpcc1000' --query='select * from  history;' --number-of-queries=1 --debug-info &
