$(function () {
    $('#key_word_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/key-word-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "org_id": $.trim($('#select_org_id_search').val()),
                        "key_word": $.trim($('#key_word').val()),
                        "check_mod": $.trim($('#check_mod').val()),
                        "exclude_words": $.trim($('#exclude_words').val()),
                        "search_time": $.trim($('#search_time').val()),
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
                        var strData = '<input type="checkbox" class="checkchild" value="' + row['id'] + '"/>';
                        return strData;
                    }
                },
                {
                    data: "index",
                    title: "序号",
                    width: "5%"
                },
                {
                    data: "org_id",
                    title: "组织名称",
                    width: "8%",
                },
                {
                    data: "key_word", title: "关键词", width: "8%",
                },
                {
                    data: 'search_time', title: '检索日期', width: '8%',
                },
                {
                    data: 'exclude_words', title: '过滤词', width: '20%',
                },
                {data: 'check_mod', title: '检索模式', width: '8%'},
                {data: 'count', title: '检索数量', width: '8%'},
                {
                    title: "操作",
                    width: "8%",
                    "render": function (data, type, row, meta) {
                        const strDelete = "<a class=\"btn btn-sm btn-danger\" href=javascript:delete_key_word(\"" + row["id"] + "\") role=\"button\" title=\"Delete\"><i class=\"fa fa-trash-o\"></i></a>";
                        return strDelete;
                    }
                }
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
        $("#key_word_table").DataTable().draw(true);
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete('#key_word_table', '/key-word-delete');
    });
    $("#checkbox_cron_task_xscan").click(function () {
        if (this.checked) {
            $("#input_cron_rule_xscan").prop("disabled", false);
            $("#input_cron_comment_xscan").prop("disabled", false);
            $("#label_cron_rule_xscan").prop("disabled", false);
        } else {
            $("#input_cron_rule_xscan").prop("disabled", true);
            $("#input_cron_comment_xscan").prop("disabled", true);
            $("#label_cron_rule_xscan").prop("disabled", true);
        }
    })
    $('#select_poc_type_xscan').change(function () {
        load_pocfile_list(true, false, $('#select_poc_type_xscan').val())
    });
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

//新建关键词窗口
$("#create_key_word").click(function () {
    $('#new_key_word').modal('toggle');
});

//新建任务创建
$("#create_key_word_task").click(function () {
    $('#new_key_word_task').modal('toggle');
    load_pocfile_list(true, false, "default")
});
$('#select_poc_type_xscan').change(function () {
    load_pocfile_list(true, false, $('#select_poc_type_xscan').val())
});
$("#start_xscan_task").click(function () {
    const formData = new FormData();
    if ($('#select_org_id_task_xscan').val() === "") {
        swal('Warning', '必须选择要执行任务的组织！', 'error');
        return
    }
    let cron_rule = "";
    if ($('#checkbox_cron_task_xscan').is(":checked")) {
        cron_rule = $('#input_cron_rule_xscan').val();
        if (!cron_rule) {
            swal('Warning', '请输入定时任务规则', 'error');
            return;
        }
    }
    formData.append("xscan_type", "xfofa");
    formData.append("org_id", $('#select_org_id_task_xscan').val());
    formData.append("fingerprint", $('#checkbox_fingerpint_xscan').is(":checked"));
    formData.append("xraypoc", $('#checkbox_xraypoc_xscan').is(":checked"));
    formData.append("xraypocfile", $('#select_poc_type_xscan').val() + '|' + $('#input_xray_poc_file_xscan').val());


    formData.append("nucleipoc", $('#checkbox_nucleipoc_xscan').is(":checked"));
    formData.append("nucleipocfile", $('#input_nuclei_poc_file_xscan').val());

    formData.append("taskcron", $('#checkbox_cron_task_xscan').is(":checked"));
    formData.append("cronrule", cron_rule);
    formData.append("croncomment", $('#input_cron_comment_xscan').val());
    if (formData.get("xraypoc") === "true" && formData.get("fingerprint") === "false") {
        swal('Warning', '漏洞扫描需要开启指纹扫描步骤选项', 'error');
        return;
    }
    if (formData.get("nucleipoc") === "true" && formData.get("fingerprint") === "false") {
        swal('Warning', '漏洞扫描需要开启指纹扫描步骤选项', 'error');
        return;
    }
    $.ajax({
        url: '/task-start-xscan',
        type: 'POST',
        cache: false,
        data: formData,
        processData: false,
        contentType: false
    }).done(function (res) {
        if (res['status'] == "success") {
            swal({
                    title: "新建任务成功！",
                    text: res['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                },
                function () {
                    $('#new_key_word_task').modal('hide');
                });
        } else {
            swal('Warning', '新建任务失败！' + res['msg'], 'error');
        }
    }).fail(function (res) {
        swal('Warning', '新建任务失败！' + res['msg'], 'error');
    });
});

$("#start_add_key_search_word").click(function () {
    var formData = new FormData();
    if ($('#select_import_org_id_task').val() == "") {
        swal('Warning', '必须选择归属组织！', 'error');
        return
    }
    formData.append("add_key_word", $('#add_key_word').val());
    formData.append("add_exclude_words", $('#add_exclude_words').val());
    formData.append("add_search_time", $('#add_search_time').val());
    formData.append("add_check_mod", $('#add_check_mod').val());
    formData.append("add_count", $('#add_count').val());
    formData.append("add_org_id", $('#select_import_org_id_task').val());

    $.ajax({
        url: '/key-word-add',
        type: 'POST',
        cache: false,
        data: formData,
        processData: false,
        contentType: false
    }).done(function (res) {
        if (res['status'] == "success") {
            swal({
                    title: "导入成功！",
                    text: res['msg'],
                    type: "success",
                    confirmButtonText: "确定",
                    confirmButtonColor: "#41b883",
                    closeOnConfirm: true,
                },
                function () {
                    $('#new_key_word').modal('hide');
                });
        } else {
            swal('Warning', '导入失败！' + res['msg'], 'error');
        }
    }).fail(function (res) {
        swal('Warning', '导入失败！' + res['msg'], 'error');
    });
});

function delete_key_word(id) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除当前关键词配置信息，请确认！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/key-word-del",
                {
                    "id": id,
                }, function (data, e) {
                    if (e === "success") {
                        $('#key_word_table').DataTable().draw(false);
                    }
                });
        });
}