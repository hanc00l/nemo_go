$(function () {
    $.urlParam = function (param) {
        const results = new RegExp(`[?&]${param}=([^&#]*)`).exec(window.location.href);
        return results ? results[1] : null;
    };
    if (isNotEmpty($.urlParam('unitName'))) {
        $('#unit_name').val(decodeURIComponent($.urlParam('unitName')));
    }
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
                "url": "/icp-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "domain": $('#domain').val(),
                        "unit_name": $('#unit_name').val(),
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
                {data: "domain", title: "域名", width: "15%"},
                {
                    data: "unit_name", title: "主体名称", width: "30%",
                    render: function (data, type, row) {
                        let strData = '';
                        strData += data + '&nbsp;<a href=https://www.beianx.cn/search/' + data + ' target="_blank"><img src="/static/images/beianx-favicon.ico" style="max-width:16px;" title="在beianx.cn中查看"></a>';
                        strData += '<a href=/unit-list?unitName=' + encodeURIComponent(data) + ' target="_blank">&nbsp;<i class="fa fa-external-link" aria-hidden="true"></i></a>';
                        return strData;
                    }
                },
                {data: "company_type", title: "主体性质", width: "8%"},
                {data: "service_licence", title: "备案号", width: "15%"},
                {data: "verify_time", title: "审核日期", width: "8%"},
                {data: "update_time", title: "更新时间", width: "8%"},
                {
                    title: "操作",
                    width: "6%",
                    "render": function (data, type, row, meta) {
                        let strData = ''
                        strData += '&nbsp;<a onclick=delete_unit("' + row.id + '") href="#"><i class="fa fa-trash" title="删除"></i></a>';
                        return strData;
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
    $('#online_search').click(function () {
        $('#search_key').val("");
        $('#online_search_modal').modal('toggle');
    });
    $("#online_api_search").click(function () {
        const key = $("#search_key").val();
        if (!isNotEmpty(key)) {
            swal('Warning', '查询的关键字不能为空', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/icp-online-api-search',
            data: {
                key: key,
            },
            success: function (data) {
                show_response_message("在线查询", data, function () {
                    $("#list_table").DataTable().draw(false);
                    $('#online_search_modal').modal('hide');
                })
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete('#list_table', '/icp-delete');
    });
})
;

function delete_unit(id) {
    delete_by_id('#list_table', '/icp-delete', id);
}
