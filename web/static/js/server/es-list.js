$(function () {
    //$('#btnsiderbar').click();
    $('#es_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/es-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "query": $('#query').val(),
                    });
                }
            },
            columns: [
                {
                    data: "id",
                    width: "5%",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    "render": function (data, type, row) {
                        const strData = '<input type="checkbox" class="checkchild" value="' + row['id'] + '"/>';
                        return strData;
                    }
                },
                {
                    data: "index",
                    title: "序号",
                    width: "5%",
                },
                {
                    data: "host",
                    title: "Host",
                    width: "6%",
                    render: function (data, type, row, meta) {
                        let strData;
                        let ipData = data.split(":");
                        if (isIpv4(ipData[0]) || isIpv6(ipData[0])) strData = '<a href="/ip-info?workspace=' + row['workspace'] + '&&ip=' + ipData[0] + '" target="_blank">' + data + '</a>';
                        else strData = '<a href="/domain-info?workspace=' + row['workspace'] + '&&domain=' + data + '" target="_blank">' + data + '</a>';
                        if (row["header"] !== "") {
                            strData += '<br/><a href="javascript:show_http_content(\'' + row["workspace_guid"] + '\',\'' + row["id"] + '\')"><i class="fa fa-file-code-o" title="网站正文"></i></a>';
                        }
                        return strData;
                    }
                },
                {
                    data: "ip", title: "IP", width: "10%",
                    render: function (data, type, row, meta) {
                        let strData = '';
                        strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data;
                        if (row["status"] !== "") strData += '<br/>[' + row["status"] + ']</div>';
                        return strData;
                    }
                },

                {
                    data: "location", title: "Location", width: "8%",
                    render: function (data, type, row, meta) {
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                    }
                },
                {
                    data: "title", title: "Title", width: "15%",
                    render: function (data, type, row, meta) {
                        let strData = "";
                        for (let i in row['iconimage']) {
                            let icon_hash_file = row['iconimage'][i].split("|");
                            strData += '<img src=/webfiles/' + row['workspace_guid'] + '/iconimage/' + icon_hash_file[1] + ' width="24px" height="24px" title="icon_hash=' + icon_hash_file[0] + '"/>&nbsp;';
                        }
                        if (strData.length > 0) strData += "<br/>";
                        strData += data;
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + strData + '</div>';

                    }
                },
                {
                    data: 'header', title: 'Header', width: '25%',
                    render: function (data, type, row, meta) {
                        if (row['header'] === "" && row["cert"] === "" && row["service"] === "" && row['banner'] === "") return "";
                        let is_active = "active";
                        let strData = ' <ul class="nav nav-tabs" id="myTab_' + row["id"] + '">';
                        if (row['header'] !== "") {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#header_' + row["id"] + '">Header</a></li>'
                            is_active = "";
                        }
                        if (row['cert'] !== "") {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#cert_' + row["id"] + '">Cert</a></li>'
                            is_active = "";
                        }
                        if (row['service'] !== "") {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#service_' + row["id"] + '">Service</a></li>'
                            is_active = "";
                        }
                        if (row['banner'] !== "") {
                            strData += ' <li class="nav-item active"><a class="nav-link ' + is_active + '" data-toggle="tab" href="#banner_' + row["id"] + '">Banner</a></li>'
                        }
                        strData += '</ul>';
                        is_active = "show active";
                        strData += '<div class="tab-content" id="myTabContent_' + row["id"] + '">';
                        if (row['header'] !== "") {
                            strData += '<div id="header_' + row["id"] + '" class="tab-pane fade ' + is_active + '"> ';//<div class="card card-body">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">'
                            strData += '<pre>';
                            strData += html2Escape(row["header"]) + '</pre></div></div>';//</div>';
                            is_active = "";
                        }
                        if (row['cert'] !== "") {
                            strData += '<div id="cert_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">'
                            strData += row["cert"] + '</div></div>';
                            is_active = "";
                        }
                        if (row['service'] !== "") {
                            strData += '<div id="service_' + row["id"] + '" class="tab-pane fade ' + is_active + '">' + row["service"] + '</div>';
                            is_active = "";
                        }
                        if (row['banner'] !== "") {
                            strData += '<div id="banner_' + row["id"] + '" class="tab-pane fade ' + is_active + '">';
                            strData += '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                            strData += html2Escape(row["banner"]) + '</div></div>';
                        }
                        return strData;
                    }
                },
                {
                    data: "screenshot", title: "ScreenShot", width: "10%",
                    "render": function (data, type, row, meta) {
                        let title = '';
                        let index = 0;
                        let host = ''
                        let hostArray = row['host'].split(":");
                        if (hostArray.length >= 1) {
                            host = hostArray[0]
                        } else {
                            return
                        }
                        for (let i in data) {
                            index++;
                            if (index >= 4) {
                                break;
                            }
                            let thumbnailFile = data[i].replace('.png', '_thumbnail.png');
                            let imgTitle = data[i].replace(".png", "").replace("_", ":");
                            title += '<img src="/webfiles/' + row['workspace_guid'] + '/screenshot/' + host + '/' + thumbnailFile + '" class="img"  style="margin-bottom: 5px;margin-left: 5px;" title="' + imgTitle + '" onclick="show_bigpic(\'/webfiles/' + row['workspace_guid'] + '/screenshot/' + host + '/' + data[i] + '\')"/>'
                        }
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + title + '</div>';
                    }
                },
                {
                    data: 'update_time', title: '更新时间', width: '8%'
                },
            ],
            infoCallback: function (settings, start, end, max, total, pre) {
                return "共<b>" + total + "</b>条记录，当前显示" + start + "到" + end + "记录";
            },
            drawCallback: function (setting) {
                var _this = $(this);
                var tableId = _this.attr('id');
                var pageDiv = $('#' + tableId + '_paginate');
                pageDiv.append(
                    '<i class="fa fa-arrow-circle-o-right fa-lg" aria-hidden="true"></i><input id="' + tableId + '_gotoPage" type="text" style="height:20px;line-height:20px;width:40px;"/>' +
                    '<a class="paginate_button" aria-controls="' + tableId + '" tabindex="0" id="' + tableId + '_goto">Go</a>')
                $('#' + tableId + '_goto').click(function (obj) {
                    var page = $('#' + tableId + '_gotoPage').val();
                    var thisDataTable = $('#' + tableId).DataTable();
                    var pageInfo = thisDataTable.page.info();
                    if (isNaN(page)) {
                        $('#' + tableId + '_gotoPage').val('');
                        return;
                    } else {
                        var maxPage = pageInfo.pages;
                        var page = Number(page) - 1;
                        if (page < 0) {
                            page = 0;
                        } else if (page >= maxPage) {
                            page = maxPage - 1;
                        }
                        $('#' + tableId + '_gotoPage').val(page + 1);
                        thisDataTable.page(page).draw('page');
                    }
                })
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
        $("#es_table").DataTable().draw(true);
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete('#es_table', '/es-delete');
    });
    // workspace
    get_user_workspace_list();
    $('#select_workspace').change(function () {
        change_user_workspace('#ip_table');
    });
    // img
    $('.imgPreview').click(function () {
        $('.imgPreview').hide();
    });
    // 查询后按回车键触发
    $('#query').bind("keydown", function (e) {
        var theEvent = e || window.event;
        var code = theEvent.keyCode || theEvent.which || theEvent.charCode;
        if (code == 13) {
            $('#search').trigger("click");
            return false;
        }
    });
    $('ul a:first').tab('show');
});

/**
 * 移除 dataTables默认参数，并设置分页值
 * @param param
 */
function init_dataTables_defaultParam(param) {
    for (var key in param) {
        if (key.indexOf("columns") == 0 || key.indexOf("order") == 0 || key.indexOf("search") == 0) { //以columns开头的参数删除
            delete param[key];
        }
    }
    param.pageSize = param.length;
    param.pageNum = (param.start / param.length) + 1;
}

//  获取用户的工作空间
function get_user_workspace_list() {
    document.getElementById('li_workspace').style.visibility = 'visible';
    $.post("/workspace-user-list", function (data) {
        $("#select_workspace").empty();
        for (let i = 0; i < data.WorkspaceInfoList.length; i++) {
            $("#select_workspace").append("<option value='" + data.WorkspaceInfoList[i].workspaceId + "'>" + data.WorkspaceInfoList[i].workspaceName + "</option>")
        }
        $('#select_workspace').val(data.CurrentWorkspace);
    });
}


// 手工切换用户的工作空间
function change_user_workspace(dataTableId) {
    let newWorkspace = $("#select_workspace").val();
    $.post("/workspace-user-change", {"workspace": newWorkspace}, function (data) {
        if (data['status'] == 'success') {
            swal({
                    title: "切换工作空间成功！",
                    text: data['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                },
                function () {
                    $(dataTableId).DataTable().draw(false);
                });
        } else {
            swal('Warning', "切换工作空间失败! " + data['msg'], 'error');
        }
    });
}

function html2Escape(sHtml) {
    var temp = document.createElement("div");
    (temp.textContent != null) ? (temp.textContent = sHtml) : (temp.innerText = sHtml);
    var output = temp.innerHTML.replace(/\"/g, "&quot;").replace(/\'/g, "&acute;");
    temp = null;
    //output = output
    return output;
}

/**
 * 图片放大显示
 * @param src
 */
function show_bigpic(src) {
    $('.imgPreview img').attr('src', src);
    $('.imgPreview').show();
}


//批量删除
function batch_delete(dataTableId, url) {
    swal({
            title: "确定要批量删除选定的目标?",
            text: "该操作会删除所有选定目标的所有信息！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $(dataTableId).DataTable().$('input[type=checkbox]:checked').each(function (i) {
                let id = $(this).val().split("|")[0];
                $.ajax({
                    type: 'post',
                    async: false,
                    url: url + '?id=' + id,
                    success: function (data) {
                    },
                    error: function (xhr, type) {
                    }
                });
            });
            $(dataTableId).DataTable().draw(false);
        });
}

// 显示网站正文
function show_http_content(workspace, id) {
    $('#text_content_http').val("");
    let url = "es-get-body?workspace=" + workspace + "&&id=" + id;
    $.post(url,
        function (data, e) {
            if (e === "success" && data['status'] == 'success') {
                $('#text_content_http').val(data['msg']);
            }
        })
    $('#showHttpInfo').modal('toggle');
}

function show_query_expr() {
    $('#showSearchExpr').modal('toggle');
}