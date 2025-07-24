$(function () {
    $.urlParam = function (param) {
        const results = new RegExp(`[?&]${param}=([^&#]*)`).exec(window.location.href);
        return results ? results[1] : null;
    };
    if (isNotEmpty($.urlParam('unitName'))) {
        $('#unit_name').val(decodeURIComponent($.urlParam('unitName')));
    }
    if (isNotEmpty($.urlParam('parentUnitName'))) {
        $('#parent_unit_name').val(decodeURIComponent($.urlParam('parentUnitName')));
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
                "url": "/unit-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "parent_unit_name": $('#parent_unit_name').val(),
                        "unit_name": $('#unit_name').val(),
                        "fuzzy": $('#checkbox_fuzzy').prop('checked'),
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
                    data: "parent_unit_name", title: "上级主体名称", width: "20%",
                    render: function (data, type, row) {
                        if (isNotEmpty(data)) return data + '<a href=/unit-list?unitName=' + encodeURIComponent(data) + '>&nbsp;<i class="fa fa-external-link" aria-hidden="true"></i></a>';
                        else return "";
                    }
                },
                {
                    data: "unit_name", title: "主体名称", width: "20%",
                    render: function (data, type, row) {
                        let strData = '';
                        strData += data + '&nbsp;<a href=https://www.riskbird.com/ent/' + data + '.html?entid=' + row["ent_id"] + ' target=blank><img src="/static/images/riskbird-favicon.ico" style="max-width:16px;" title="在RiskBird中查看"></a>'
                        strData += '<a href=/unit-list?parentUnitName=' + encodeURIComponent(data) + '>&nbsp;<i class="fa fa-external-link" aria-hidden="true"></i></a>';

                        return strData;
                    }
                },
                {data: "status", title: "状态", width: "8%"},
                {data: "type", title: "类型", width: "8%"},
                {data: "invest_ration", title: "投资比例", width: "8%"},
                {
                    data: "icp_data", title: "ICP备案", width: "10%",
                    "render": function (data, type, row) {
                        if (data > 0) {
                            return '<a href=/icp-list?unitName=' + encodeURIComponent(row.unit_name) + ' target=blank>备案信息:' + data + '</a>';
                        } else {
                            return "";
                        }
                    }
                },
                {data: "update_time", title: "更新时间", width: "8%"},
                {
                    title: "操作",
                    width: "6%",
                    "render": function (data, type, row, meta) {
                        let strData = ''
                        strData += '&nbsp;<a onclick=edit_unit("' + row.id + '") href="#"><i class="fa fa-edit" title="修改"></i></a>';
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
        $('#online_search_modal').modal('toggle');
    });
    $("#online_api_search").click(function () {
        const unit_name = $("#search_unit_name").val();
        const is_branch = $('#search_is_branch').prop('checked');
        const is_invest = $('#search_is_invest').prop('checked');
        const is_icp_online = $('#search_icp_online_api').prop('checked');
        const invest_ration = parseInt($('#search_invest_ration').val());
        const max_depth = parseInt($('#search_max_depth').val());
        const cookie = $('#search_online_api_cookie').val();
        if (!isNotEmpty(unit_name)) {
            swal('Warning', '名称不能为空', 'error');
            return;
        }
        if (!isNotEmpty(cookie)) {
            swal('Warning', 'cookie不能为空', 'error');
            return;
        }
        if (isNaN(invest_ration) || invest_ration < 0 || invest_ration > 100) {
            swal('Warning', '投资比例必须为0-100之间的整数', 'error');
            return;
        }
        if (isNaN(max_depth) || max_depth < 1 || max_depth > 3) {
            swal('Warning', '最大深度必须为1-3之间的整数', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/unit-online-api-search',
            data: {
                unit_name: unit_name,
                is_branch: is_branch,
                is_invest: is_invest,
                is_icp_online: is_icp_online,
                invest_ration: invest_ration,
                max_depth: max_depth,
                cookie: cookie
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
    $('#update_unit_info').click(function () {
        update_unit();
    });
    $('#unit_export').click(function () {
        unit_export();
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete('#list_table', '/unit-delete');
    });
})
;

function delete_unit(id) {
    delete_by_id('#list_table', '/unit-delete', id);
}

function edit_unit(id) {
    $('#edit_modal').modal('toggle');
    $.ajax({
        type: 'post',
        url: '/unit-get',
        data: {id: id},
        success: function (data) {
            $('#hidden_id').val(data.id);
            $('#edit_parent_unit_name').val(data.parent_unit_name);
            $('#edit_unit_name').val(data.unit_name);
            $('#edit_unit_type').val(data.type);
            $('#edit_invest_ration').val(data.invest_ration);
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}

function update_unit() {
    if (!isNotEmpty($('#edit_unit_name').val())) {
        swal('Warning', '名称不能为空', 'error');
        return;
    }
    if ($('#edit_unit_type').val() === "invest") {
        if (!isNotEmpty($('#edit_invest_ration').val())) {
            swal('Warning', '投资比例不能为空', 'error');
            return;
        }
        if ($('#edit_invest_ration').val() < 0 || $('#edit_invest_ration').val() > 100) {
            swal('Warning', '投资比例必须为0-100之间的整数', 'error');
            return;
        }
    }
    $.ajax({
        type: 'post',
        url: '/unit-update',
        data: {
            id: $('#hidden_id').val(),
            parent_unit_name: $('#edit_parent_unit_name').val(),
            unit_name: $('#edit_unit_name').val(),
            type: $('#edit_unit_type').val(),
            invest_ration: $('#edit_invest_ration').val()
        },
        success: function (data) {
            $('#edit_modal').modal('hide');
            $('#list_table').DataTable().draw(true);
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}


function unit_export() {
    $.ajax({
        type: 'POST',
        url: '/unit-export',
        data: {
            "parent_unit_name": $('#parent_unit_name').val(),
            "unit_name": $("#unit_name").val(),
            "fuzzy": $('#checkbox_fuzzy').prop('checked'),
        },
        dataType: 'json',
        success: function (data) {
            if (data['status'] === 'success') {
                const url = data['msg'];
                const a = document.createElement('a');
                a.href = url;
                a.download = 'unit.csv';
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
