import dateutil.parser
import time
import logging
from datetime import datetime
from configparser import SafeConfigParser

import dataset
import requests
from dateutil.parser import parse

try:
    config = SafeConfigParser()
    config.read('config.cfg')
except:
    raise
else:
    log_level = config.get('logging', 'level')

logging_format = '%(asctime)s %(levelname)s %(message)s'
if log_level == 'info':
    logging_level = logging.INFO
elif log_level == 'warn':
    logging_level = logging.WARNING
elif log_level == 'debug':
    logging_level = logging.DEBUG

logging.basicConfig(format=logging_format, level=logging.INFO)
requests.packages.urllib3.disable_warnings()


class runner:
    def fetch(self):
        url = f'https://opendata.epa.gov.tw/api/v1/AQI?format=json&ts={int(time.time())}'
        req = requests.get(url, verify=False)
        if req.status_code == 200:
            try:
                req.json()
            except Exception as e:
                logging.exception(e)
            else:
                self.record(req.json())

    def record(self, rec):
        db = dataset.connect(config.get('db', 'url')))
        table=db['OpenDataAir']

        def clean(data_):
            tmp=[]
            for data in data_:
                data.update(
                    {'PM25': data['PM2.5'], 'PM25_AVG': data['PM2.5_AVG']})
                data.pop('PM2.5')
                data.pop('PM2.5_AVG')
                tmp.append(data)
            return tmp

        now=datetime.now()
        # result = table.find_one(PublishTime=datetime(now.year, now.month, now.day, now.hour))
        for post in clean(rec):
            if post['PublishTime'] == now.strftime('%Y-%m-%d %H:00'):
                try:
                    result=table.find_one(
                        SiteName = post['SiteName'], PublishTime = now.strftime('%Y-%m-%d %H:00'))
                except Exception as e:
                    raise
                    logging.critical('QueryDatabase Error')
                else:
                    if result:
                        logging.info('Pass...')
                    else:
                        try:
                            table.insert(post)
                            logging.info('Run...')
                        except Exception as e:
                            logging.info('error')
                            logging.exception(e)


tester=runner()

while True:
    tester.fetch()
    time.sleep(180)
