$(function () {
    $('#profile_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/profile-list",
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
                    data: "name", title: "任务配置名称", width: "30%",
                    "render": function (data, type, row, meta) {
                        return '<a href="profile-edit?id=' + row.id + '">' + data + '</a>';
                    }
                },
                {data: "executors", title: "任务配置项", width: "20%"},
                {
                    data: "status", title: "状态", width: "5%",
                    "render": function (data, type, row, meta) {
                        if (data === "disable") {
                            return '<span class="badge badge-secondary">Disable</span>';
                        } else {
                            return '<span class="badge badge-success">Enable</span>';
                        }

                    }
                },
                {data: "sort_number", title: "排序", width: "8%"},
                {data: "create_time", title: "创建时间", width: "10%"},
                {data: "update_time", title: "更新时间", width: "10%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        return '<a onclick=delete_profile("' + row.id + '") href="#"><i class="fa fa-trash"></i><span>Delete</span></a>';
                    }
                }
            ]
        }
    );//end datatable


    $(".checkall").click(function () {
        const check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });
});

function delete_profile(id) {
    delete_by_id('#profile_table','/profile-delete',id);
}
