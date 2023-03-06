$(function () {
    $('#btnsiderbar').click();
    load_org_list();
    // //获取任务的状态信息
    get_task_status();
    setInterval(function () {
        get_task_status();
    }, 60 * 1000);
    //列表全选
    $(".checkall").click(function () {
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });
    $('[data-toggle="tooltip"]').tooltip();
    $('.imgPreview').click(function () {
        $('.imgPreview').hide();
    });
    // 鼠标悬停时，更新任务状态信息
    $('#span_show_task').popover({
        trigger: 'manual',
        placement: 'bottom',
        html: true,
        container: false,//'#span_show_task',
        content: update_task_status,
    }).on('mouseenter', function () {
        var _this = this;
        $(this).popover('show');
        $(this).siblings('.popover').on('mouseleave', function () {
            $(_this).popover('hide');
        });
    }).on('mouseleave', function () {
        var _this = this;
        setTimeout(function () {
            if (!$('.popover:hover').length) {
                $(_this).popover('hide')
            }
        }, 100);
    });
    $('#nav_tabs').on('shown.bs.tab', function (event) {
        if (event.target.innerText === "漏洞验证") {
            load_pocfile_list();
        }
    })
});

/**
 * 更新正在进行的任务信息
 */
function update_task_status() {
    get_task_status();
    let info = "进行中/待运行/已完成任务数";
    $.ajax({
        type: 'post',
        async: false,
        url: "/dashboard-task-started-info",
        success: function (data) {
            info = "";
            for (let i = 0; i < data.length; i++) {
                info += "<li class=\"list-group-item\"><strong>" + data[i].task_name + "-></strong>" + data[i].task_args + " <strong>@</strong>" + data[i].task_starting + "</li>";
            }
        },
        error: function (xhr, type) {
        }
    });
    if (info == "") info = "进行中/待运行/已完成任务数";
    return info;
}

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

/**
 * 加载组织列表
 */
function load_org_list() {
    $("#select_org_id_search").append("<option value=''>--全部--</option>")
    $("#select_org_id_task").append("<option value=''>--无--</option>")
    $("#select_batchscan_org_id_task").append("<option value=''>--无--</option>")
    $("#select_import_org_id_task").append("<option value=''>--无--</option>")
    $("#select_org_id_task_xscan").append("<option value=''>--无--</option>")
    $.post("/org-getall", {}, function (data, e) {
        if (e === "success") {
            for (let i = 0; i < data.length; i++) {
                $("#select_org_id_search").append("<option value='" + data[i].id + "'>" + data[i].name + "</option>")
                $("#select_org_id_task").append("<option value='" + data[i].id + "'>" + data[i].name + "</option>")
                $("#select_batchscan_org_id_task").append("<option value='" + data[i].id + "'>" + data[i].name + "</option>")
                $("#select_import_org_id_task").append("<option value='" + data[i].id + "'>" + data[i].name + "</option>")
                $("#select_org_id_task_xscan").append("<option value='" + data[i].id + "'>" + data[i].name + "</option>")
            }
            $('#select_org_id_search').val($('#hidden_org_id').val());
            $('#select_org_id_task').val($('#hidden_org_id').val());
            $('#select_batchscan_org_id_task').val($('#hidden_org_id').val());
            $('#select_import_org_id_task').val($('#hidden_org_id').val());
            $('#select_org_id_task_xscan').val($('#hidden_org_id').val());
        }
    });
}

/**
 * 加载poc文件列表
 */
function load_pocfile_list(xray = true, nuclei = true, xray_type = "default") {
    if (xray) {
        $.post("/vulnerability-load-xray-pocfile",
            {
                "type": xray_type
            },
            function (data, e) {
                if (e === "success") {
                    $("#datalist_xray_poc_file").empty();
                    $("#datalist_xray_poc_file_xscan").empty();
                    for (let i = 0; i < data.length; i++) {
                        $("#datalist_xray_poc_file").append("<option value='" + data[i] + "'>" + data[i] + "</option>")
                        $("#datalist_xray_poc_file_xscan").append("<option value='" + data[i] + "'>" + data[i] + "</option>")
                    }
                }
            });
    }
    if (nuclei) {
        $.post("/vulnerability-load-nuclei-pocfile", {}, function (data, e) {
            if (e === "success") {
                $("#datalist_nuclei_poc_file").empty();
                for (let i = 0; i < data.length; i++) {
                    $("#datalist_nuclei_poc_file").append("<option value='" + data[i] + "'>" + data[i] + "</option>")
                }
            }
        });
    }
}

/**
 * 获取任务状态
 */
function get_task_status() {
    $.post("/dashboard-task-info", function (data) {
        $("#span_show_task").html(data['task_info']);
    });
}

/**
 * html字符转义
 * @param sHtml
 * @returns {string}
 */
function html2Escape(sHtml) {
    var temp = document.createElement("div");
    (temp.textContent != null) ? (temp.textContent = sHtml) : (temp.innerText = sHtml);
    var output = temp.innerHTML.replace(/\"/g, "&quot;").replace(/\'/g, "&acute;");
    temp = null;

    return output;
}

/**
 * 获取选中的Tab索引号
 * 0: portscan
 * 1: vulverify
 */
function getCurrentTabIndex(tab_id) {
    var $tabs = $(tab_id).children('li');
    var i = 0;
    $tabs.each(function () {
        var $tab = $(this);
        if ($tab.children('a').hasClass('active')) {
            return false;
        } else {
            i++;
        }
    });
    return i;
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


this.encodeHtml = function (s) {
    let REGX_HTML_ENCODE = /"|&|'|<|>|[\x00-\x20]|[\x7F-\xFF]|[\u0100-\u2700]/g;
    s = (s != undefined) ? s : this.toString();
    return (typeof s != "string") ? s :
        s.replace(REGX_HTML_ENCODE,
            function ($0) {
                var c = $0.charCodeAt(0), r = ["&#"];
                c = (c == 0x20) ? 0xA0 : c;
                r.push(c);
                r.push(";");
                return r.join("");
            });
};

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