# 在Nemo中集成和使用飞书知识库

## v1.0

2024-01-20

## 一. 飞书知识库介绍

飞书知识库是飞书的一项功能，可以用来存储和管理团队的知识，类似于企业内部的wiki。飞书知识库可通过客户端或网页，在团队进行在线协作使用。同时飞书知识库支持markdown语法，可以方便的编辑和查看。

飞书知识库的缺点是没有离线使用的功能，需要联网才能使用。

飞书知识库的使用方法可以参考[飞书官方文档](https://www.feishu.cn/hc/zh-CN/articles/360024984973)。

## 二. 为什么要把飞书知识库集成到Nemo中

在很多项目或工作场景中，团队的沟通和协作一般都是通过微信、钉钉或飞书等工具进行的，即时沟通平台的优势是可以快速的进行沟通和协作，但是缺点是信息的存储和管理比较混乱，不方便后续的查找、使用以及整理报告和归档。像语雀、飞书知识库等在线文档平台就可以很好的满足需求。

由于飞书提供了可以自定义企业组织架构的功能，可以方便的创建团队和项目，同时飞书知识库支持markdown语法，可以方便的进行团队协作的编辑、查看和导出，因此把飞书知识库集成到Nemo中，可以方便的在Nemo中查看和使用飞书知识库中的内容。

## 三. 在Nemo中使用飞书知识库的准备工作

### 1. 注册飞书帐号、创建新企业，邀请团队成员加入企业

### 2. 在飞书中创建知识库，并授予团队成员知识管的管理员或编辑权限

![feishu-201.png](image%2Ffeishu-201.png)

![feishu-202.png](image%2Ffeishu-202.png)

### 3. 在飞书开发者后台中创建自建应用

+ 在飞书开发者后台中创建自建应用后，创建一个发行版本，应用即可使用并获得应用的AppId和AppSecret。

![feishu-311.png](image%2Ffeishu-311.png)

![feishu-312.png](image%2Ffeishu-312.png)

![feishu-313.png](image%2Ffeishu-313.png)

![feishu-314.png](image%2Ffeishu-314.png)

+ 根据自建应用调用的API，为自建应用开通知识库API使用的权限。调用的API清单如下：

  *认证及授权*

    - 获取登录预授权码：GET /open-apis/authen/v1/index。应用请求用户身份验证时，需构造登录链接，并引导用户跳转至此链接。用户登录成功后会生成登录预授权码
      code，并作为参数追加到重定向URL。
    - 获取 user_access_token：POST /open-apis/authen/v1/oidc/access_token。根据登录预授权码 code 获取 user_access_token。
    - 刷新 user_access_token：POST /open-apis/authen/v1/oidc/refresh_access_token。user_access_token 的最大有效期是
      2小时左右。当 user_access_token 过期时，可以调用本接口获取新的 user_access_token。

  *知识库*

    - 获取知识空间子节点列表：GET /open-apis/wiki/v2/spaces/:space_id/nodes。此接口用于分页获取Wiki节点的子节点列表。
      此接口为分页接口。由于权限过滤，可能返回列表为空，但分页标记（has_more）为true，可以继续分页请求。
    - 获取知识空间节点信息：GET /open-apis/wiki/v2/spaces/get_node。获取知识空间节点信息。
    - 创建知识空间节点：POST /open-apis/wiki/v2/spaces/:space_id/nodes。此接口用于在知识节点里创建节点到指定位置。
    - 获取文档纯文本内容：GET /open-apis/docx/v1/documents/:document_id/raw_content。获取文档的纯文本内容。

  *导出*

    - 创建导出任务：POST
      /open-apis/drive/v1/export_tasks。创建导出任务，将云文档导出为指定格式的本地文件，目前支持新版文档、电子表格、多维表格和旧版文档。该接口为异步接口，任务创建完成即刻返回，并不会阻塞等待到任务执行成功，因此需要结合查询导出任务结果接口获取导出结果。
    - 查询导出任务结果：GET /open-apis/drive/v1/export_tasks/:
      ticket。根据创建导出任务返回的ticket轮询导出任务的结果，通过本接口获取到导出产物的文件token之后，可调用下载导出文件接口将导出产物下载到本地。
    - 下载导出文件：GET /open-apis/drive/export_tasks/file/:file_token/download。根据查询导出任务结果返回的导出产物token，下载导出产物文件到本地。

+ 需开通的权限如下：

    - 获取登录预授权码：无需权限
    - 获取 user_access_token：无需权限
    - 刷新 user_access_token：无需权限
    - 获取知识空间子节点列表：查看、编辑和管理知识库，查看知识库
    - 获取知识空间节点信息：查看、编辑和管理知识库，查看知识库
    - 创建知识空间节点：查看、编辑和管理知识库
    - 获取文档纯文本内容：创建及编辑新版文档，查看新版文档
    - 创建导出任务：导出云文档
    - 查询导出任务结果：导出云文档
    - 下载导出文件：导出云文档

  ![feishu-321.png](image%2Ffeishu-321.png)

  ![feishu-322.png](image%2Ffeishu-322.png)

  ![feishu-323.png](image%2Ffeishu-323.png)

+ 获取知识库的space_id

  在飞书知识库中，查看知识库的“分享知识库”，在分享链接中可以看到知识库的space_id。注意：不建议对外分享，这里只是为了获取space_id。

  ![feishu-331.png](image%2Ffeishu-331.png)

+ 在自建应用中，设置重定向URL

  设置重定向URL，用于获取登录预授权码code。对于Nemo来说，重定向URL的地址为：`https://<nemo_server:port>/wiki-feishu-code`
  。如果没有开启TLS加密需将https改为http。

  ![feishu-341.png](image%2Ffeishu-341.png)

## 四. 在Nemo中使用飞书知识库

### 1. 在Nemo中配置飞书自建应用的AppId和AppSecret，并获取用户访问权限Token

- 在“Config-配置管理-知识库：飞书自建应用设置”中，将飞书自建应用的AppId和AppSecret填入，并点击“从飞书用户访问权限Token”按钮。

![feishu-411.png](image%2Ffeishu-411.png)

- 输入Nemo的IP地址，点击“前往飞书验证”，获取用户访问权限Token。如果获取成功，将会在页面中显示“获取用户AccessToken获取成功。

![feishu-412.png](image%2Ffeishu-412.png)

![feishu-413.png](image%2Ffeishu-413.png)

![feishu-414.png](image%2Ffeishu-414.png)

### 2. 在Nemo中将知识库与工作空间进行关联

在“System-工作空间”中，选择要关联的工作空间，点击“Edit”，在“知识库的SpaceId”中设置知识库的Space_id，点击“更新”。

![feishu-421.png](image%2Ffeishu-421.png)

### 3. 在Nemo中查看和使用知识库

#### 同步知识库

点击“同步知识库”按钮，从飞书知识库中同步知识库的文档列表到Nemo中。

如果文档在Nemo中不存在，则在Nemo中创建文档，如果文档在Nemo中存在，则更新文档的标题和时间。如果文档在飞书知识库中不存在，则会从Nemo中删除文档。

Nemo会调用飞书知识库的API，获取每一个文档的纯文本内容，并提取出其中的IP和域名，如果该文档中的IP和域名在Nemo中存在，则会创建文档和IP、域名的关联关系。在IP或域名的列表及详情页面中，可以查看到该IP或域名在哪些文档中出现过。

#### 新建文档

通过Nemo新建一个文档，点击“新建文档”按钮，即可在Nemo中新建一个文档，并通过API同步在飞书的知识库中创建文档。创建文档时，标题将同步到飞书知识库中的文档标题，备注只是用于在Nemo中使用。

![feishu-431.png](image%2Ffeishu-431.png)

#### 编辑文档

点击文档列表中的标题，即可打开网页版的飞书知识库的文档编辑页面。文档的编辑最终是由飞书云文档在线进行的，Nemo只是提供了一个打开文档编辑页面的链接。

![feishu-432.png](image%2Ffeishu-432.png)

#### 修改文档

修改文档只是修改了文档的标题和备注，不会修改飞书知识库中的文档信息。同时可以选择是否将文档在Nemo中置顶（不会影响飞书知识库中的文档），如果文档已导出到Nemo，也可以选择是否清除导出的文档。

![feishu-433.png](image%2Ffeishu-433.png)

#### 导出文档

点击“Export”按钮，即可将文档导出到Nemo中。导出的文档将会保存在Nemo的“知识库”目录下。已导出的文档在标题前会有标记，点击即可下载到本地进行查看和编辑。

![feishu-434.png](image%2Ffeishu-434.png)