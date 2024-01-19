$(function () {
    //$('#btnsiderbar').click();
    $('#document_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/wiki-list",
                "type": "post",
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
                {
                    data: "index", title: "序号", width: "5%",
                    "render": function (data, type, row, meta) {
                        let strData = data;
                        if (row["pin_index"] === 1) strData = '<i class="fa fa-thumb-tack fa-rotate-90" style="color: orange" title="已置顶"></i>';
                        return strData;
                    }
                },
                {
                    data: "title", title: "标题", width: "30%",
                    "render": function (data, type, row, meta) {
                        let str = '';
                        if (row['export'] !== '') str += "<a href='" + row['export'] + "' target='_blank' title='已导出的文档'><i class='fa fa-file-text'></i></a>&nbsp;";
                        str += '<a href="https://nqmgg3xrjzi.feishu.cn/wiki/' + row['node_token'] + '" target="' + row['node_token'] + '">' + row['title'] + '</a>';
                        return str;
                    }
                },
                {data: "comment", title: "备注", width: "25%"},
                {data: "update_time", title: "更新时间", width: "10%"},
                {data: "create_time", title: "创建时间", width: "15%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        let str = '<a onclick="edit_document(' + row.id + ')" role="button" data-toggle="modal" href="#" title="Edit" data-target="#editdocument"><i class="fa fa-pencil"></i><span>Edit</span></a>';
                        str += '&nbsp;<a onclick="export_document(' + row.id + ')" role="button" href="#" title="Export"><i class="fa fa-cloud-download"></i><span>Export</span></a>';
                        return str;
                    }
                },
            ],
            infoCallback: function (settings, start, end, max, total, pre) {
                return "共<b>" + total + "</b>条记录，当前显示" + start + "到" + end + "记录";
            },
        },
    );//end datatable
    $(".checkall").click(function () {
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });

    $("#button_sync").click(function () {
        $.post("/wiki-sync",
            {}, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
                    swal({
                            title: "同步知识成功",
                            text: data['msg'],
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            $("#document_table").DataTable().draw(false);
                        });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    });

    $("#document_update").click(function () {
        let document_id = $('#document_id').val()
        if (!document_id) return;

        $.post("/wiki-update?id=" + document_id,
            {
                "comment": $('#document_comment').val(),
                "pin_index": $('#checkbox_pin_to_top').prop('checked') ? 1 : 0,
                "remove_exported_file": $('#checkbox_remove_exported_file').is(':checked'),
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
                    swal({
                            title: "更新文档成功",
                            text: data['msg'],
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            $("#document_table").DataTable().draw(false);
                            $('#editdocument').modal('hide');
                        });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    });
});

/**
 * 编辑文档
 * @param id
 */
function edit_document(id) {
    $.ajax({
        type: 'post',
        url: '/wiki-get?id=' + id,
        dataType: 'json',
        success: function (e) {
            const data = eval(e);
            $('#document_id').val(data.id);
            $('#document_title').val(data.title);
            $('#document_comment').val(data.comment);
            $('#checkbox_pin_to_top').prop('checked', data.pin_index > 0);
            $('#checkbox_remove_exported_file').prop('checked', false);
        },
        error: function (xhr, type) {
        }
    });
}

/**
 * 导出文档
 * @param id
 */
function export_document(id) {
    $.post("/wiki-export?id=" + id,
        {}, function (data, e) {
            if (e === "success" && data['status'] === 'success') {
                swal({
                        title: "导出文档成功",
                        text: data['msg'],
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    },
                    function () {
                        $("#document_table").DataTable().draw(false);
                    });
            } else {
                swal('Warning', data['msg'], 'error');
            }
        });
}