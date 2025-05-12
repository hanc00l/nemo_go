$(function () {
    $.urlParam = function (param) {
        const results = new RegExp(`[?&]${param}=([^&#]*)`).exec(window.location.href);
        return results ? results[1] : null;
    };
    load_select_list('/org-all-list', $('#select_org_id_search'));
    // 从 URL 中获取authority参数，并设置到查询条件中；用于其它页面跳转到资产查询页面时，自动带上查询条件
    let authority = $.urlParam('authority');
    if (isNotEmpty(authority)) {
        $('#query').val('authority=="' + authority + '"');
    }
    $('#asset_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/asset-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "task_id": $.urlParam('task_id'),
                        "query": $('#query').val(),
                        "new": $('#checkbox_asset_new').is(":checked"),
                        "update": $('#checkbox_asset_update').is(":checked"),
                        "order_by_date": $('#checkbox_select_order_by_date').is(":checked"),
                    });
                }
            },
            columns: [
                {
                    data: "id",
                    width: "5%",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    render: function (data, type, row) {
                        return '<input type="checkbox" class="checkchild" value="' + row['id'] + '"/>';
                    }
                },
                {
                    data: "index",
                    title: "序号",
                    width: "5%",
                    render: function (data, type, row, meta) {
                        if (row["honeypot"] === true) return "<span style='color:red;font-weight:bold' title='host匹配到自定义蜜罐列表'>蜜罐</span>";
                        else return data;
                    }
                },
                {
                    data: "authority",
                    title: "资产-组织",
                    width: "12%",
                    render: function (data, type, row, meta) {
                        let strData = "";
                        strData = generateLink(data, row["service"]);
                        let brOrg = "</br>";
                        if (row['org']) {
                            strData += brOrg + '<span class="custom-shape custom-shape-org" title="所属组织">' + row['org'] + '</span>';
                            brOrg = "&nbsp;";
                        }
                        if (row['icp_company']) {
                            strData += brOrg + '<span class="custom-shape custom-shape-org" title="ICP备案">' + row['icp_company'] + '</span>';
                        }
                        let br = "</br>";
                        if (row["new"] || row["update"]) {
                            strData += br + '<i class="fa  fa-spinner"  style="color:#FFA500" aria-hidden="true" title="新增或有更新"></i>&nbsp;';
                            br = "";
                        }
                        if (row["vul"]) {
                            strData += br + '<i class="fa fa-bolt" style="color:red" aria-hidden="true" title="有漏洞信息"></i>&nbsp;';
                            br = "";
                        }
                        if (row["cdn"]) {
                            strData += br + '<i class="fa fa-cloud" style="color:#FFA500" aria-hidden="true" title="CDN"></i>&nbsp;';
                            br = "";
                        }
                        if (row['memo']) {
                            strData += br + '<i class="fa fa-flag" style="color:red" title="有备忘信息"></i>&nbsp;';
                            br = "";
                        }
                        if (row["header"] !== "") {
                            strData += br + '<a href="javascript:show_http_content(\'' + row["id"] + '\')"><i class="fa fa-file-code-o" title="网站正文"></i>&nbsp;</a>';
                            br = "";
                        }
                        return strData;
                    }
                },
                {
                    data: "ip", title: "IP-归属地", width: "10%",
                    render: function (data, type, row, meta) {
                        let strData = '';
                        strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data;
                        if (row['location'] !== null) {
                            strData += '</br>'
                            for (let i in row['location']) {
                                strData += '<span class="text-wrap"><small>' + row['location'][i] + '</small></span>&nbsp;';
                            }
                        }
                        strData += '</div>';
                        return strData;
                    }
                },
                {
                    data: "port", title: "端口-协议", width: "8%", render: function (data, type, row, meta) {
                        let strData = data;
                        if (row['service'] !== "" && row['service'] !== 'unknown') {
                            strData += '<span class="badge badge-light text-dark">' + row['service'] + '</span>';
                        }
                        if (row["status"] !== "") {
                            let color = "black"; // 默认颜色
                            let statusText = row["status"];

                            if (statusText.startsWith("1")) {
                                color = "#ADD8E6"; // 1xx: 浅蓝色
                            } else if (statusText.startsWith("2")) {
                                color = "#4CAF50"; // 2xx: 绿色
                            } else if (statusText.startsWith("3")) {
                                color = "#FFD700"; // 3xx: 黄色
                            } else if (statusText.startsWith("4")) {
                                color = "#FFA500"; // 4xx: 橙色
                            } else if (statusText.startsWith("5")) {
                                color = "#FF0000"; // 5xx: 红色
                            }
                            strData += `<span style="color: ${color};">[${statusText}]</span>`;
                        }

                        return strData;
                    }
                },
                {
                    data: "title", title: "标题-应用", width: "12%",
                    render: function (data, type, row, meta) {
                        let strData = "";
                        for (let i in row['iconimage']) {
                            strData += '<img src="data:image/x-icon;base64,';
                            strData += row['iconimage'][i];
                            strData += '" style="max-width:32px;"/>&nbsp;';
                        }
                        if (strData.length > 0) strData += "<br/>";
                        strData += encodeHtml(data);
                        if (row["app"]) {
                            strData += "</br>";
                            for (let i in row['app']) {
                                strData += '<span class="custom-shape custom-shape-app" title="应用">' + row['app'][i] + '</span>&nbsp;';
                            }
                        }
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + strData + '</div>';

                    }
                },
                {
                    data: 'header', title: '指纹信息', width: '26%',
                    render: function (data, type, row, meta) {
                        if (!isNotEmpty(row['header']) && !isNotEmpty(row['cert']) && !isNotEmpty(row['banner']) && !isNotEmpty(row['vul']) && !isNotEmpty(row['memo']) && !isNotEmpty(row['icp']) && !isNotEmpty(row['whois'])) return "";
                        let is_active = "active";
                        let strData = ' <ul class="nav nav-tabs" id="myTab_' + row["id"] + '">';
                        if (isNotEmpty(row['header'])) {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#header_' + row["id"] + '">Header</a></li>'
                            is_active = "";
                        }
                        if (isNotEmpty(row['cert'])) {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#cert_' + row["id"] + '">Cert</a></li>'
                            is_active = "";
                        }
                        if (isNotEmpty(row['banner'])) {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#banner_' + row["id"] + '">Banner</a></li>'
                            is_active = "";
                        }
                        if (isNotEmpty(row['vul'])) {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#vul_' + row["id"] + '">Vul</a></li>'
                            is_active = "";
                        }
                        if (isNotEmpty(row['memo'])) {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#memo_' + row["id"] + '">Memo</a></li>'
                            is_active = "";
                        }
                        if (isNotEmpty(row['icp'])) {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#icp_' + row["id"] + '">ICP</a></li>'
                            is_active = "";
                        }
                        if (isNotEmpty(row['whois'])) {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#whois_' + row["id"] + '">Whois</a></li>'
                        }
                        strData += '</ul>';
                        is_active = "show active";
                        strData += '<div class="tab-content" id="myTabContent_' + row["id"] + '">';
                        if (isNotEmpty(row['header'])) {
                            strData += '<div id="header_' + row["id"] + '" class="tab-pane fade ' + is_active + '"> ';//<div class="card card-body">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">'
                            strData += '<pre>';
                            strData += encodeHtml(row["header"]) + '</pre></div></div>';
                            is_active = "";
                        }
                        if (isNotEmpty(row['cert'])) {
                            strData += '<div id="cert_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<a class="btn btn-outline-light text-dark" data-toggle="collapse" href="#collapse_cert_' + row["id"] + '" role="button" aria-expanded="false" aria-controls="collapseExample">' +
                                '+ Certificate</a>';
                            strData += '<div class="collapse" id="collapse_cert_' + row["id"] + '"><div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;"><pre>' + row['cert'] + '</pre></div>';
                            strData += '</div></div>';
                            is_active = "";
                        }
                        if (isNotEmpty(row['banner'])) {
                            strData += '<div id="banner_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                            strData += encodeHtml(row["banner"]) + '</div></div>';
                            is_active = "";
                        }
                        if (isNotEmpty(row['vul'])) {
                            strData += '<div id="vul_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                            strData += row['vul'] + '</div></div>';
                            is_active = "";
                        }
                        if (isNotEmpty(row['memo'])) {
                            strData += '<div id="memo_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                            strData += encodeHtml(row['memo']) + '</div></div>';
                            is_active = "";
                        }
                        if (isNotEmpty(row['icp'])) {
                            strData += '<div id="icp_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                            strData += encodeHtml(row['icp']) + '</div></div>';
                            is_active = "";
                        }
                        if (isNotEmpty(row['whois'])) {
                            strData += '<div id="whois_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                            strData += encodeHtml(row['whois']) + '</div></div>';
                        }
                        return strData;
                    }
                },
                {
                    data: "screenshot", title: "Web截屏", width: "10%",
                    "render": function (data, type, row, meta) {
                        let title = '';
                        let index = 0;
                        let host = row['host'];
                        for (let i in data) {
                            index++;
                            if (index >= 4) {
                                break;
                            }
                            let thumbnailFile = data[i].replace('.png', '_thumbnail.png');
                            let imgTitle = data[i].replace(".png", "").replace("_", ":");
                            title += '<img src="/webfiles/' + row['workspace'] + '/screenshot/' + host + '/' + thumbnailFile + '" class="img"  style="margin-bottom: 5px;margin-left: 5px;" title="' + imgTitle + '" onclick="show_bigpic(\'/webfiles/' + row['workspace'] + '/screenshot/' + host + '/' + data[i] + '\')"/>'
                        }
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + title + '</div>';
                    }
                },
                {
                    data: 'update_time', title: '更新时间', width: '6%'
                },
            ],
            infoCallback: function (settings, start, end, max, total, pre) {
                return "共<b>" + total + "</b>条记录，当前显示" + start + "到" + end + "记录";
            },
            drawCallback: function (setting) {
                get_statistic_data();
                set_page_jump($(this));
            }
        }
    );//end datatable
    $(".checkall").click(function () {
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });
    $('[data-toggle="tooltip"]').tooltip();
    //搜索
    $("#search").click(function () {
        $('#statistic_port').html("");
        $('#statistic_service').html("");
        $('#statistic_server').html("");
        $('#statistic_title').html("");
        $('#statistic_icon').html("");
        const activeTab = getActiveTab();
        if (activeTab) {
            if (activeTab.id === "query2") {
                $('#query').val(get_query_expr());
            }
        }
        $("#asset_table").DataTable().draw(true);
    });
    //批量删除
    $("#batch_delete").click(function () {
        let url = "/asset-delete";
        let task_id = $.urlParam('task_id')
        if (task_id !== null && task_id !== "") url += "?task_id=" + task_id;
        batch_delete('#asset_table', url);
    });
    $("#show_memo").click(function () {
        let ids = $('#asset_table').DataTable().$('input[type=checkbox]:checked');
        if (ids.length > 0) {
            show_memo_content(ids[0].value);
        } else {
            alert("请选择要操作一项资产，如果选择了多个资产，只会显示第一个资产的备忘信息。");
        }
    });
    $('#btn_saveMemo').click(function () {
        let ids = $('#asset_table').DataTable().$('input[type=checkbox]:checked');
        if (ids.length > 0) {
            save_memo_contnet(ids[0].value);
        }
    });
    $('#asset_export').click(function () {
        asset_export();
    });
    $("#block_ip").click(function () {
        block_asset();
    });
    // img
    $('.imgPreview').click(function () {
        $('.imgPreview').hide();
    });
    // 查询后按回车键触发
    $('#query').bind("keydown", function (e) {
        const theEvent = e || window.event;
        const code = theEvent.keyCode || theEvent.which || theEvent.charCode;
        if (code === 13) {
            $('#search').trigger("click");
            return false;
        }
    });
    $('ul a:first').tab('show');
});


function get_query_expr(expr) {
    const query_expr_list = [];
    if ($('#select_org_id_search').val().length > 0) {
        query_expr_list.push('org=="' + $('#select_org_id_search').val() + '"');
    }
    if ($('#host').val().length > 0) {
        query_expr_list.push('host=="' + $('#host').val() + '"');
    }
    if ($('#ip').val().length > 0) {
        query_expr_list.push('ip=="' + $('#ip').val() + '"');
    }
    if ($('#port').val().length > 0) {
        query_expr_list.push('port=="' + $('#port').val() + '"');
    }
    if ($('#title').val().length > 0) {
        query_expr_list.push('title=="' + $('#title').val() + '"');
    }
    if ($('#fingerprint').val().length > 0) {
        const fp = $('#fingerprint').val();
        const finger_expr = '(server=="' + fp + '" || service=="' + fp + '" || banner=="' + fp + '" || app=="' + fp + '" || cert=="' + fp + '" || header=="' + fp + '")';
        query_expr_list.push(finger_expr);
    }
    if (query_expr_list.length > 0) {
        return query_expr_list.join(' && ');
    }
    return "";
}

/**
 * 图片放大显示
 * @param src
 */
function show_bigpic(src) {
    $('.imgPreview img').attr('src', src);
    $('.imgPreview').show();
}

// 显示网站正文
function show_http_content(id) {
    $('#text_content_http').val("");
    let url = "asset-http-content?id=" + id;
    let task_id = $.urlParam('task_id')
    if (task_id !== null && task_id !== "") url += "&&task_id=" + task_id;

    $.ajax({
        type: 'POST',
        url: url,
        dataType: 'json',
        success: function (response) {
            $('#text_content_http').val(response['msg']);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
    $('#showHttpInfo').modal('toggle');
}

// 显示备注
function show_memo_content(id) {
    $('#text_content_memo').val("");
    let url = "asset-memo-content?id=" + id;
    let task_id = $.urlParam('task_id')
    if (task_id !== null && task_id !== "") url += "&&task_id=" + task_id;
    $.ajax({
        type: 'POST',
        url: url,
        dataType: 'json',
        success: function (response) {
            $('#text_content_memo').val(response['msg']);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
    $('#showMemoInfo').modal('toggle');
}

function save_memo_contnet(id) {
    let url = "asset-memo-update?id=" + id;
    let task_id = $.urlParam('task_id')
    if (task_id !== null && task_id !== "") url += "&&task_id=" + task_id;
    let memo = $('#text_content_memo').val();
    $.ajax({
        type: 'POST',
        url: url,
        data: {memo: memo},
        dataType: 'json',
        success: function (response) {
            show_response_message("更新备忘信息成功", response, function () {
                $('#showMemoInfo').modal('hide');
                $("#asset_table").DataTable().draw(false);
            })
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function show_query_expr() {
    $('#showSearchExpr').modal('toggle');
}

function getActiveTab() {
    // 通过 ID 获取 Tab 列表
    const tabList = document.getElementById('searchTypeTab');
    // 获取当前选中的 Tab（通过 active 类）
    return tabList.querySelector('.nav-link.active'); // 返回当前选中的 Tab 元素
}

function generateLink(input, service) {
    if (isNotEmpty(input)) {
        if (service === "http" || service === "https") {
            const url = service + "://" + input;
            return `${input}<a href="${url}" target="_blank">&nbsp;<i class="fa fa-external-link" aria-hidden="true"></i></a>`;
        }
    }
    // 正则表达式，用于检测是否为IPv4地址
    const ipv4Regex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;

    // 检查是否包含端口号
    const hasPort = input.includes(":");
    let url;

    if (hasPort) {
        // 分割 host 和 port
        const [host, port] = input.split(":");
        // 判断端口号是否以 443 结尾
        if (port.endsWith("443")) {
            url = `https://${input}`;
        } else {
            url = `http://${input}`;
        }
    } else {
        // 检查是否为IPv4地址
        if (ipv4Regex.test(input)) {
            url = `http://${input}`;
        } else if (input.includes(".")) {
            // 如果包含点但不是IPv4地址，假设为域名
            url = `https://${input}`;
        } else {
            // 其他情况（如localhost或纯数字）
            url = `http://${input}`;
        }
    }

    // 生成 <a> 超链接
    return `${input}<a href="${url}" target="_blank">&nbsp;<i class="fa fa-external-link" aria-hidden="true"></i></a>`;
}

function get_statistic_data() {
    $.ajax({
        type: 'POST',
        url: '/asset-statistic',
        data: {
            "task_id": $.urlParam('task_id'),
            "query": $('#query').val(),
            "new": $('#checkbox_asset_new').is(":checked"),
            "update": $('#checkbox_asset_update').is(":checked"),
            "order_by_date": $('#checkbox_select_order_by_date').is(":checked"),
        },
        dataType: 'json',
        success: function (response) {
            $('#statistic_port').html(get_result_output(response['port']));
            $('#statistic_service').html(get_result_output(response['service']));
            $('#statistic_app').html(get_result_output(response['app']));
            $('#statistic_title').html(get_result_output(response['title']));
            $('#statistic_icon').html(get_result_output_for_icon(response['icon_hash_bytes']));
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function get_result_output(arrayObj, limit_length = 40) {
    if (arrayObj == null || arrayObj.length === 0) return "";

    let output = "";
    let output_prefix = "<br>";
    for (let index = 1; index <= arrayObj.length; index++) {
        const key = arrayObj[index - 1]["Field"];
        const value = arrayObj[index - 1]["Count"];
        output += output_prefix;
        output += '<span class="badge badge-pill badge-info">' + value + '</span>';
        if (value < 10) {
            output += "&nbsp;";
        }
        if (limit_length > 0) {
            let showed_key = encodeHtml(String(key).substr(0, limit_length));
            if (String(key).length > limit_length) showed_key += '...';
            output += showed_key;
        } else {
            output += key;
        }
    }
    return output
}


function get_result_output_for_icon(arrayObj) {
    if (arrayObj == null || arrayObj.length === 0) return "";

    let output = "";
    let output_prefix = "<br>";
    for (let index = 1; index <= arrayObj.length; index++) {
        const key = arrayObj[index - 1]["Field"];
        const value = arrayObj[index - 1]["Count"];
        output += output_prefix;
        output += '<span class="badge badge-pill badge-info">' + value + '</span>';
        if (value < 10) {
            output += "&nbsp;";
        }
        output += '<img src="data:image/x-icon;base64,';
        output += key;
        output += '" style="max-width:32px;"/>&nbsp;';
    }
    return output
}


function asset_export() {
    const activeTab = getActiveTab();
    if (activeTab) {
        if (activeTab.id === "query2") {
            $('#query').val(get_query_expr());
        }
    }
    $.ajax({
        type: 'POST',
        url: '/asset-export',
        data: {
            "task_id": $.urlParam('task_id'),
            "query": $('#query').val(),
            "new": $('#checkbox_asset_new').is(":checked"),
            "update": $('#checkbox_asset_update').is(":checked"),
            "order_by_date": $('#checkbox_select_order_by_date').is(":checked"),
        },
        dataType: 'json',
        success: function (data) {
            if (data['status'] === 'success') {
                const url = data['msg'];
                const a = document.createElement('a');
                a.href = url;
                a.download = 'asset.csv';
                a.click();
            } else {
                alert(data['msg']);
            }
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function block_asset() {
    swal({
            title: "确定要一键拉黑选定的资产HOST吗?",
            text: "该操作会将资产的host加入到黑名单列表中，同时从数据库中删除要关的host的资产记录！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $('#asset_table').DataTable().$('input[type=checkbox]:checked').each(function (i) {
                let id = $(this).val().split("|")[0];
                $.ajax({
                    type: 'POST',
                    url: '/asset-block',
                    data: {
                        "id": id,
                        "task_id": $.urlParam('task_id'),
                    },
                    dataType: 'json',
                    success: function (data) {
                    },
                    error: function (xhr, type) {
                    }
                });
            });
            $('#asset_table').DataTable().draw(false);
        });
}