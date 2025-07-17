$(function () {
    $.urlParam = function (param) {
        const results = new RegExp(`[?&]${param}=([^&#]*)`).exec(window.location.href);
        return results ? results[1] : null;
    };
    $('#list_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/vul-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "host": $('#host').val(),
                        "pocfile": $('#pocfile').val(),
                        "source": $('#source').val(),
                        "severity": $('#severity').val(),
                        "task_id": $.urlParam('task_id'),
                    });
                }
            },
            columns: [
                {
                    data: "id",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    "render": function (data, type, row) {
                        return '<input type="checkbox" class="checkchild" value="' + data + '"/>';
                    }
                },
                {data: "index", title: "序号", width: "5%"},
                {
                    data: "authority", title: "资产", width: "10%",
                    "render": function (data, type, row) {
                        return '<a href="/asset-list?authority=' + data + '" target=_blank>' + data + '</a>';
                    }
                },
                {data: "name", title: "名称", width: "35%"},
                {
                    data: "pocfile", title: "Poc文件", width: "20%",
                    "render": function (data, type, row) {
                        return '<a href="/vul-info?id=' + row["id"] + '" target=_blank>' + data + '</a>';
                    }
                },
                {
                    data: "severity", title: "等级", width: "8%",
                    "render": function (data, type, row) {
                        return get_severity_color(data);
                    }
                },
                {data: "create_time", title: "创建时间", width: "8%"},
                {data: "update_time", title: "更新时间", width: "8%"},
                {
                    title: "操作",
                    width: "6%",
                    "render": function (data, type, row, meta) {
                        if (!isNotEmpty(row.task_id)) return '&nbsp;<a onclick=delete_vul("' + row.id + '") href="#"><i class="fa fa-trash" title="删除漏洞"></i></a>';
                        else return "";
                    }
                }
            ],
            infoCallback: function (settings, start, end, max, total, pre) {
                return "共<b>" + total + "</b>条记录，当前显示" + start + "到" + end + "记录";
            },
            drawCallback: function (setting) {
                set_page_jump($(this));
            }
        }
    );//end datatable

    $(".checkall").click(function () {
        const check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });
    $('#search').click(function () {
        $("#list_table").DataTable().draw(true);
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete('#list_table', '/vul-delete');
    });
});

function delete_vul(id) {
    delete_by_id('#list_table', '/vul-delete', id);
}

function get_severity_color(severity) {
    switch (severity) {
        case "critical":
            return '<span class="badge bg-danger" style="font-size: 12px;color: white;">' + severity + '</span>';
        case "high":
            return '<span class="badge bg-danger" style="font-size: 12px;color: white;">' + severity + '</span>';
        case "medium":
            return '<span class="badge bg-warning" style="font-size: 12px;color: white;">' + severity + '</span>';
        case "low":
            return '<span class="badge bg-success" style="font-size: 12px;color: white;">' + severity + '</span>';
        case "info":
            return '<span class="badge bg-info" style="font-size: 12px;color: white;">' + severity + '</span>';
        default:
            return '<span class="badge bg-secondary" style="font-size: 12px; color: white;">' + severity + '</span>';
    }
}
