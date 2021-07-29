#!/usr/bin/env python3
#coding: utf-8
#
# Created by hancool for nemo_go 2021.7.12.
# This .py file execute pocsuite and save result in json format.
#

import argparse
import json
from urllib.parse import urlparse

from pocsuite3.api import get_results
from pocsuite3.api import init_pocsuite
from pocsuite3.api import start_pocsuite


class poc():
    def __init__(self):
        self.poc_file = ""
        self.url = ""
        self.outuput_file = ""
        self.THREADS = 10
        self.url_file = ""

    def __pocsuite3_scanner(self, _poc_config):
        '''
        Pocsuite3 API调用
        '''
        init_pocsuite(_poc_config)
        start_pocsuite()
        result = get_results()

        return result

    def __parse_pocsuite3_result(self, results):
        '''
        解析执行结果
        '''
        vul_results = []
        for data in results:
            if data.status == 'success':
                pr = urlparse(data.url)
                r = {'source': "pocsuite3", 'target': pr.hostname,
                     'poc_file': self.poc_file}
                if 'VerifyInfo' in data.result and 'URL' in data.result['VerifyInfo']:
                    r['url'] = data.result['VerifyInfo']['URL']
                else:
                    r['url'] = data.url
                r['extra'] = ''
                if 'extra' in data.result:
                    if len(str(data.result['extra'])) > 2000:
                        r['extra'] = str(data.result['extra'])[:2000] + '...'
                    else:
                        r['extra'] = str(data.result['extra'])

                vul_results.append(r)

        return vul_results

    def __write_output(self, vul_results):
        '''
        保存JSON结果到文件中
        '''
        with open(self.outuput_file, "w") as f:
            f.write(json.dumps(vul_results))

    def execute_verify(self):
        '''
        调用pocsuite3进行漏洞验证
        '''
        _poc_config = {
            'poc': self.poc_file,
            'threads': self.THREADS,
            'quiet': False,
            'random_agent': True,
        }
        if self.url_file:
            _poc_config['url_file'] = self.url_file
        else:
            _poc_config['url'] = self.url

        results = self.__pocsuite3_scanner(_poc_config)
        vul_results = self.__parse_pocsuite3_result(results)
        self.__write_output(vul_results)


def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("-u", "--url",  help="url")
    parser.add_argument("-f", "--url_file",  help="url list")
    parser.add_argument("-r", "--poc_file", required=True,
                        help="poc file name")
    parser.add_argument("--threads", default="10", help="threads")
    parser.add_argument("-o", "--output_file", required=True,
                        help="ouput result josn filename")
    args = parser.parse_args()
    if args.url == "" and args.url_file == "":
        print("the following arguments are required: -u/--url or -f/--url_file")
        exit(0)
    return args


def main():
    args = get_args()

    p = poc()
    p.poc_file = args.poc_file
    p.url = args.url
    p.url_file = args.url_file
    p.THREADS = args.threads
    p.outuput_file = args.output_file

    p.execute_verify()


if __name__ == "__main__":
    main()
