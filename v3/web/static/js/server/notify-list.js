$(function () {
    $('#notify_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/notify-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {});
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
                    data: "name",
                    title: "通知名称", width: "35%",
                    "render": function (data, type, row, meta) {
                        return '<a href=notify-update?id=' + row.id + '><span> ' + data + '</span></a>';
                    }
                },
                {data: "category", title: "通知类型", width: "8%"},
                {
                    data: "status", title: "状态", width: "8%",
                    "render": function (data, type, row, meta) {
                        if (data === "disable") {
                            return '<span class="badge badge-secondary">Disable</span>';
                        } else {
                            return '<span class="badge badge-success">Enable</span>';
                        }

                    }
                },
                {data: "sort_number", title: "排序", width: "8%"},
                {data: "create_time", title: "创建时间", width: "12%"},
                {data: "update_time", title: "更新时间", width: "12%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        return '<a onclick=delete_notify("' + row.id + '") href="#"><i class="fa fa-trash"></i><span>删除</span></a>';
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
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });

});

function delete_notify(id) {
    delete_by_id('#notify_table', '/notify-delete', id);
}
