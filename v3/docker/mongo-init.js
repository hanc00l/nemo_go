// MongoDB 初始化脚本
// 创建数据库和集合，并插入初始数据

// 1. 创建 nemoSystem 数据库并初始化数据
db = db.getSiblingDB('nemoSystem');

// 创建 user 表并插入数据
db.createCollection('user');
db.user.insertOne({
    "_id": ObjectId("67947fee59943f160d21fc23"),
    "username": "nemo",
    "password": "0a595da98826241197130c0eb1efc86defc841f8651efa2157d03a75c1c9769d8e2dc80ead506df3b8adb9c7933345e18f08cd1b1311d560a6cc16a8fd78e4b0",
    "description": "",
    "role": "superadmin",
    "sort_number": 100,
    "status": "enable",
    "workspace_id": [
        "9690c0a2-e4fc-43f0-a259-7c0d054be229"
    ],
    "create_time": ISODate("2025-01-25T06:08:46.236Z"),
    "update_time": ISODate("2025-03-13T01:32:27.972Z")
});

// 创建 workspace 表并插入数据
db.createCollection('workspace');
db.workspace.insertOne({
    "_id": ObjectId("67947fde59943f160d21fc1f"),
    "workspace_id": "9690c0a2-e4fc-43f0-a259-7c0d054be229",
    "name": "默认工作空间",
    "description": "默认工作空间",
    "status": "enable",
    "sort_number": 100,
    "create_time": ISODate("2025-01-25T06:08:30.633Z"),
    "update_time": ISODate("2025-03-15T14:25:42.252Z")
});

// 2. 创建 9690c0a2-e4fc-43f0-a259-7c0d054be229 数据库并初始化数据
db = db.getSiblingDB('9690c0a2-e4fc-43f0-a259-7c0d054be229');

// 创建 org 表并插入数据
db.createCollection('org');
db.org.insertOne({
    "_id": ObjectId("67948118ffb6b4755ec5af45"),
    "name": "测试",
    "description": "测试",
    "sort_number": 100,
    "status": "enable",
    "create_time": ISODate("2025-01-25T06:13:44.748Z"),
    "update_time": ISODate("2025-03-16T08:19:26.041Z")
});

// 创建 customData 表并插入数据
db.createCollection('customData');
db.customData.insertMany([
    {
        "_id": ObjectId("67b59523b46bc750481596fe"),
        "category": "customIPLocation",
        "description": "自定义IP支持单个IP、C段、B段及任意掩码方式，格式为：IP 归属地描述（比如192.168.3.0/24 北京-A部门）；注释以#开始 ",
        "data": "#自定义归属地\n",
        "create_time": ISODate("2025-02-19T08:24:03.931Z"),
        "update_time": ISODate("2025-02-19T08:24:03.931Z")
    },
    {
        "_id": ObjectId("67b73f8e1514b3540f686743"),
        "category": "blacklist",
        "description": "支持IP、IP段/掩码及域名；域名支持自动匹配相应的子域名；注释以#开头",
        "data": "#黑名单列表\n127.0.0.1\n.gov.cn",
        "create_time": ISODate("2025-02-20T14:43:26.849Z"),
        "update_time": ISODate("2025-02-20T14:43:26.849Z")
    },
    {
        "_id": ObjectId("67ca52206db855b68af43d04"),
        "category": "honeypot",
        "description": "格式为'ip/域名 备注'，备注为可选，以#开头为注释会被忽略；蜜罐与资产的host进行匹配。",
        "data": "#'ip/域名 备注'，备注为可选；例如：1.2.3.4 蜜罐1\n",
        "create_time": ISODate("2025-03-07T01:55:44.224Z"),
        "update_time": ISODate("2025-03-07T01:55:44.224Z")
    }
]);

// 创建 taskProfile 表并插入数据
db.createCollection('taskProfile');
db.taskProfile.insertMany([
    {
        "_id": ObjectId("67d58fa7208d24dd235ec752"),
        "name": "端口扫描(Top1000)-指纹识别",
        "description": "",
        "args": "{\"portscan\":{\"nmap\":{\"port\":\"--top-ports 1000\",\"rate\":1000,\"tech\":\"-sS\",\"maxOpenedPortPerIp\":50}},\"fingerprint\":{\"fingerprint\":{\"httpx\":true,\"fingerprintx\":true,\"screenshot\":true,\"iconhash\":true}}}",
        "sort_number": 100,
        "status": "enable",
        "create_time": ISODate("2025-03-15T14:33:11.134Z"),
        "update_time": ISODate("2025-03-15T14:33:15.347Z")
    },
    {
        "_id": ObjectId("67d58fc0208d24dd235ec75a"),
        "name": "端口扫描(Top100)-指纹识别",
        "description": "",
        "args": "{\"portscan\":{\"nmap\":{\"port\":\"--top-ports 100\",\"rate\":1000,\"tech\":\"-sS\",\"maxOpenedPortPerIp\":50}},\"fingerprint\":{\"fingerprint\":{\"httpx\":true,\"fingerprintx\":true,\"screenshot\":true,\"iconhash\":true}}}",
        "sort_number": 100,
        "status": "enable",
        "create_time": ISODate("2025-03-15T14:33:36.902Z"),
        "update_time": ISODate("2025-03-15T14:33:39.416Z")
    },
    {
        "_id": ObjectId("67d59016208d24dd235ec76f"),
        "name": "域名扫描-在线API-指纹识别",
        "description": "",
        "args": "{\"domainscan\":{\"massdns\":{\"wordlistFile\":\"normal\",\"ignorecdn\":true,\"ignorechinaother\":true,\"ignoreoutsidechina\":true,\"maxResolvedDomainPerIP\":100},\"subfinder\":{\"wordlistFile\":\"normal\",\"ignorecdn\":true,\"ignorechinaother\":true,\"ignoreoutsidechina\":true,\"maxResolvedDomainPerIP\":100}},\"onlineapi\":{\"fofa\":{\"ignorecdn\":true,\"ignorechinaother\":true,\"ignoreoutsidechina\":true,\"searchlimitcount\":2000,\"searchpagesize\":100}},\"fingerprint\":{\"fingerprint\":{\"httpx\":true,\"fingerprintx\":true,\"screenshot\":true,\"iconhash\":true}}}",
        "sort_number": 100,
        "status": "enable",
        "create_time": ISODate("2025-03-15T14:35:02.951Z"),
        "update_time": ISODate("2025-03-15T14:35:06.664Z")
    }
]);

print('MongoDB 初始化完成');