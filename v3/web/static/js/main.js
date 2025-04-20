(function () {
    "use strict";

    var treeviewMenu = $('.app-menu');

    // Toggle Sidebar
    $('[data-toggle="sidebar"]').click(function (event) {
        event.preventDefault();
        $('.app').toggleClass('sidenav-toggled');
    });

    // Activate sidebar treeview toggle
    $("[data-toggle='treeview']").click(function (event) {
        event.preventDefault();
        if (!$(this).parent().hasClass('is-expanded')) {
            treeviewMenu.find("[data-toggle='treeview']").parent().removeClass('is-expanded');
        }
        $(this).parent().toggleClass('is-expanded');
    });

    // Set initial active toggle
    $("[data-toggle='treeview.'].is-expanded").parent().toggleClass('is-expanded');

    //Activate bootstrip tooltips
    $("[data-toggle='tooltip']").tooltip();

})();

// Show response message
function show_response_message(title, data, callback) {
    if (data['status'] === 'success') {
        swal({
                title: title,
                text: data['msg'],
                type: "success",
                confirmButtonText: "确定",
                confirmButtonColor: "#41b883",
                closeOnConfirm: true,
                timer: 3000
            },
            function () {
                callback();
            });
    } else {
        swal('Warning', data['msg'], 'error');
    }
}

// Check if a value is not empty
function isNotEmpty(value) {
    // 检查 null 或 undefined
    if (value === null || value === undefined) {
        return false;
    }

    // 检查字符串是否为空
    if (typeof value === 'string') {
        return value.trim() !== "";
    }

    // 检查 number 类型是否为 0
    if (typeof value === 'number') {
        return value !== 0; // 确保 number 不为 0
    }

    // 检查数组是否为空
    if (Array.isArray(value)) {
        return value.length > 0;
    }

    // 检查对象是否为空
    if (typeof value === 'object') {
        return Object.keys(value).length > 0;
    }

    // 其他类型（如 number、boolean 等）默认不为空
    return true;
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


function encodeHtml(s) {
    // 默认值处理
    s = (s !== undefined && s !== null) ? s : "";

    // 类型检查
    if (typeof s !== "string") {
        s = String(s); // 或者 throw new Error("Expected a string");
    }

    // 定义HTML特殊字符的映射
    const htmlEntities = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#39;'
    };

    // 替换HTML特殊字符
    s = s.replace(/[&<>"']/g, (match) => htmlEntities[match]);

    // 替换其他字符为字符实体（可选）
    s = s.replace(/[^\x20-\x7E]/g, (char) => {
        return `&#${char.charCodeAt(0)};`;
    });


    return s;
}

function load_select_list(url, select, value) {
    $.ajax({
        url: url, // 服务器接口地址
        method: "POST", // 请求方法
        dataType: "json", // 期望返回的数据类型
        success: function (response) {
            // 请求成功时处理数据
            // 清空 select 的现有内容
            select.empty();
            // 添加一个默认选项（可选）
            select.append($("<option>")
                .attr("value", "")
                .text("--请选择--")
            );
            // 遍历数据并添加到 select 中
            $.each(response, function (index, item) {
                const option = $("<option>")
                    .attr("value", item.id) // 设置 value 为 item.id
                    .text(item.name);       // 设置显示文本为 item.name
                select.append(option);
            });
            if (value !== undefined) {
                select.val(value);
            }
        },
        error: function (xhr, status, error) {
            console.error("Failed to load profile data:", error);
        }
    });
}

function set_page_jump(_this) {
    const tableId = _this.attr('id');
    const pageDiv = $('#' + tableId + '_paginate');
    pageDiv.append(
        '<i class="fa fa-arrow-circle-o-right fa-lg" aria-hidden="true"></i><input id="' + tableId + '_gotoPage" type="text" style="height:20px;line-height:20px;width:40px;"/>' +
        '<a class="paginate_button" aria-controls="' + tableId + '" tabindex="0" id="' + tableId + '_goto">Go</a>')
    $('#' + tableId + '_goto').click(function (obj) {
        let page = $('#' + tableId + '_gotoPage').val();
        const thisDataTable = $('#' + tableId).DataTable();
        const pageInfo = thisDataTable.page.info();
        if (isNaN(page)) {
            $('#' + tableId + '_gotoPage').val('');
            return;
        } else {
            const maxPage = pageInfo.pages;
            page = Number(page) - 1;
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

function delete_by_id(dataTableId, url, id) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除选择的目标的所有信息！是否继续?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: 'post',
                url: url,
                data: {id: id},
                success: function (data) {
                    show_response_message("删除", data, function () {
                        $(dataTableId).DataTable().draw(false);
                    })
                },
                error: function (xhr, type, error) {
                    swal('Warning', error, 'error');
                }
            });
        });
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
                    url: url,
                    data: {id: id},
                    dataType: 'json',
                    success: function (data) {
                    },
                    error: function (xhr, type) {
                    }
                });
            });
            $(dataTableId).DataTable().draw(false);
        });
}
