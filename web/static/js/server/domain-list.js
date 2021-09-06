$(function () {
    //搜索任务
    $("#search").click(function () {
        $("#hidden_org_id").val($("#select_org_id_search").val())
        $("#domain_table").DataTable().draw(true);
    });
    //显示创建任务
    $("#create_task").click(function () {
        var checkIP = [];
        $('#domain_table').DataTable().$('input[type=checkbox]:checked').each(function (i) {
            checkIP[i] = $(this).val().split("|")[1];
        });
        $('#text_target').val(checkIP.join("\n"));
        $('#newTask').modal('toggle');
    });
    //启动任务 
    $("#start_task").click(function () {
        const target = $('#text_target').val();
        if (!target) {
            swal('Warning', '请至少输入一个Target', 'error');
            return;
        }
        if (getCurrentTabIndex() == 0) {
            $.post("/task-start-domainscan",
                {
                    "target": target,
                    'org_id': $('#select_org_id_task').val(),
                    'subdomainbrute': $('#checkbox_subdomainbrute').is(":checked"),
                    'whatweb': $('#checkbox_whatweb').is(":checked"),
                    'fld_domain': $('#checkbox_fld_domain').is(":checked"),
                    'portscan': $('#checkbox_portscan').is(":checked"),
                    'fofasearch': $('#checkbox_fofasearch').is(":checked"),
                    'networkscan': $('#checkbox_networkscan').is(":checked"),
                    'subtask': $('#checkbox_subtask').is(":checked"),
                    'subfinder': $('#checkbox_subfinder').is(":checked"),
                    'jsfinder': $('#checkbox_jsfinder').is(":checked"),
                    'httpx': $('#checkbox_httpx').is(":checked"),
                    'screenshot': $('#checkbox_screenshot').is(":checked"),
                    'icpquery': $('#checkbox_icpquery').is(":checked"),
                    'wappalyzer': $('#checkbox_wappalyzer').is(":checked"),
                }, function (data, e) {
                    if (e === "success" && data['status'] == 'success') {
                        swal({
                                title: "新建任务成功！",
                                text: "TaskId:" + data['msg'],
                                type: "success",
                                confirmButtonText: "确定",
                                confirmButtonColor: "#41b883",
                                closeOnConfirm: true,
                            },
                            function () {
                                $('#newTask').modal('hide');
                            });
                    } else {
                        swal('Warning', "添加任务失败! " + data['msg'], 'error');
                    }
                });
        } else {
            if ($('#checkbox_pocsuite3').is(":checked") == false && $('#checkbox_xray').is(":checked") == false) {
                swal('Warning', '请选择要使用的验证工具！', 'error');
                return;
            }
            if ($('#checkbox_pocsuite3').is(":checked")) {
                if ($('#input_pocsuite3_poc_file').val() == '') {
                    swal('Warning', '请选择poc file', 'error');
                    return;
                }
            }
            if ($('#checkbox_xray').is(":checked")) {
                if ($('#input_xray_poc_file').val() == '') {
                    swal('Warning', '请选择poc file', 'error');
                    return;
                }
            }
            $.post("/task-start-vulnerability", {
                "target": target,
                'pocsuite3verify': $('#checkbox_pocsuite3').is(":checked"),
                'pocsuite3_poc_file': $('#input_pocsuite3_poc_file').val(),
                'xrayverify': $('#checkbox_xray').is(":checked"),
                'xray_poc_file': $('#input_xray_poc_file').val()
            }, function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                            title: "新建任务成功！",
                            text: "TaskId:" + data['msg'],
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                        },
                        function () {
                            $('#newTask').modal('hide');
                        });
                } else {
                    swal('Warning', "添加任务失败! " + data['msg'], 'error');
                }
            });
        }
    });
    $("#domain_statistics").click(function () {
        let url = 'domain-statistics?';
        url += get_export_options();

        window.open(url);
    });
    $("#domain_memo_export").click(function () {
        var url = 'domain-memo-export?';
        url += get_export_options();

        window.open(url);
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete('#domain_table','/domain-delete');
    });

    $('#domain_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/domain-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "org_id": $('#hidden_org_id').val(),
                        "ip_address": $('#ip_address').val(),
                        "domain_address": $('#domain_address').val(),
                        "color_tag": $('#select_color_tag').val(),
                        "memo_content": $('#memo_content').val(),
                        "date_delta": $('#date_delta').val(),
                        'disable_fofa': $('#checkbox_disable_fofa').is(":checked"),
                    });
                }
            },
            columns: [
                {
                    data: "id",
                    width: "6%",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    "render": function (data, type, row) {
                        let strData = '<input type="checkbox" class="checkchild" value="' + row['id'] + "|" + row['domain'] + '"/>';
                        if (row['memo_content']) {
                            strData += '&nbsp;<span class="badge  badge-primary" data-toggle="tooltip" data-html="true" title="' + html2Escape(row['memo_content']) + '"><i class="fa fa-flag"></i></span>';
                        }
                        return strData;
                    }
                },
                {
                    data: "index", title: "序号", width: "5%",
                    "render": function (data, type, row, meta) {
                        let strData = data;
                        if (row["honeypot"].length > 0) strData = "<span style='color:red;font-weight:bold' title='" + row["honeypot"] + "'>蜜罐</span>";
                        return strData;
                    }
                },
                {
                    data: "domain",
                    title: "域名",
                    width: "12%",
                    render: function (data, type, row, meta) {
                        let strData;
                        let disable_fofa = $('#checkbox_disable_fofa').is(":checked");
                        if (row['color_tag']) {
                            strData = '<h5><a href="/domain-info?domain=' + data + '&&disable_fofa='+disable_fofa+'" target="_blank" class="badge ' + row['color_tag'] + '">' + data + '</a></h5>';
                        } else {
                            strData = '<a href="/domain-info?domain=' + data + '&&disable_fofa='+disable_fofa+'" target="_blank">' + data + '</a>';
                        }
                        if (row['vulnerability']) {
                            strData += '&nbsp;<span class="badge badge-danger" data-toggle="tooltip" data-html="true" title="' + html2Escape(row['vulnerability']) + '"><i class="fa fa-bolt"></span>';
                        }
                        return strData;
                    }
                },
                {
                    data: "ip", title: "IP地址", width: "20%",
                    "render": function (data, type, row, meta) {
                        let strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                        let pre_link = "";
                        let j = 0, len = data.length;
                        let disable_fofa = $('#checkbox_disable_fofa').is(":checked");
                        for (; j < len; j++) {
                            strData += pre_link
                            strData += '<a href="ip-info?ip=' + data[j] + '&&disable_fofa='+disable_fofa+'" target="_blank">' + data[j] + '</a>';
                            pre_link = ",";
                        }
                        strData += '</div>';
                        return strData;
                    }
                },
                {
                    data: "banner", title: "Title && Banner", width: "25%",
                    "render": function (data, type, row, meta) {
                        let title = data.substr(0, 300);
                        if (data.length > 300) title += '......';
                        const strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + title + '</div>';
                        return strData;
                    }
                },
                {
                    data: "screenshot", title: "ScreenShot", width: "20%",
                    "render": function (data, type, row, meta) {
                        let title = '';
                        for (let i in data) {
                            let thumbnailFile = data[i].replace('.png', '_thumbnail.png');
                            let imgTitle = data[i].replace(".png", "").replace("_", ":");
                            title += '<img src="screenshot/' + row['domain'] + '/' + thumbnailFile + '" class="img"  style="margin-bottom: 5px;margin-left: 5px;" title="' + imgTitle + '" onclick="show_bigpic(\'screenshot/' + row['domain'] + '/' + data[i] + '\')"/>'
                        }
                        const strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + title + '</div>';
                        return strData;
                    }
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
});

function get_export_options() {
    var url = "";
    url += 'org_id=' + encodeURI($('#select_org_id_search').val());
    url += '&ip_address=' + encodeURI($('#ip_address').val());
    url += '&domain_address=' + encodeURI($('#domain_address').val());
    url += '&color_tag=' + encodeURI($('#select_color_tag').val());
    url += '&memo_content=' + encodeURI($('#memo_content').val());
    url += '&date_delta=' + encodeURI($('#date_delta').val());
    url += '&disable_fofa=' + encodeURI( $('#checkbox_disable_fofa').is(":checked"));

    return url;
}
