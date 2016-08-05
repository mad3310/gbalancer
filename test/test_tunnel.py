#coding=utf-8
from sqlalchemy import create_engine
from sqlalchemy.sql import text
import sqlalchemy.exc as sqlexc
import time

USER='zbz'
PASSWD='zbz'
IP='127.0.0.1'
PORT='3306'
DBNAME='gbalancer'


def _test(db,i):
    #print db.execute(text('SELECT count(*) from jt_test')).fetchone()
    #print db.execute(text('call p3()')).fetchone()
    #print db.execute(text("insert into jt_test values(%d,'gbalancer')" %i))
    db.execute(text("insert into jt_test1 values(%d,'gbalancer')" %i))

def test_excute():
    i =0

    while True:
        host='mysql://%s:%s@%s:%s/%s' %(USER, PASSWD,
                IP, PORT, DBNAME)
        master_engine = create_engine(host,
            echo=False, pool_size=5, max_overflow = 10,
            pool_timeout=30)
        master_conn = master_engine.connect()
        _test(master_conn,i)
        i +=1
        print ("This is the %d time" %i)
        master_conn.close()
        time.sleep(1)

if __name__ == "__main__":
    test_excute()

