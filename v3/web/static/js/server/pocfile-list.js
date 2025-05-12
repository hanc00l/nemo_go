$(document).ready(function () {
    let currentPath = "";

    // 初始化文件列表
    loadFileList("");

    // 上级目录按钮
    $("#btnGoUp").click(function () {
        if (currentPath) {
            let pathParts = currentPath.split('/');
            pathParts.pop();
            let newPath = pathParts.join('/');
            loadFileList(newPath);
        }
    });

    // 新建目录按钮
    $("#btnNewFolder").click(function () {
        $("#newFolderModal").modal('show');
    });

    // 创建目录
    $("#btnCreateFolder").click(function () {
        let folderName = $("#folderName").val().trim();
        if (!folderName) {
            showAlert("请输入目录名称", "error");
            return;
        }

        $.ajax({
            url: "pocfile-createFolder",
            type: "POST",
            data: {
                path: currentPath,
                name: folderName
            },
            success: function (response) {
                if (response.status === "success") {
                    $("#newFolderModal").modal('hide');
                    loadFileList(currentPath);
                    showAlert("目录创建成功", "success");
                } else {
                    showAlert(response.msg, "error");
                }
            },
            error: function () {
                showAlert("请求失败", "error");
            }
        });
    });

    // 上传文件按钮
    $("#btnUpload").click(function () {
        $("#fileInput").click();
    });

    // 文件选择
    $("#fileInput").change(function () {
        if (this.files.length === 0) return;

        let formData = new FormData();
        for (let i = 0; i < this.files.length; i++) {
            let file = this.files[i];
            let ext = file.name.split('.').pop().toLowerCase();
            if (['yaml', 'yml', 'json', 'txt', 'xml','csv', 'js', 'svg'].indexOf(ext) === -1) {
                showAlert("只能上传YAML/JSON/TXT/XML/CSV/JS/SVG格式文件", "error");
                return;
            }
            formData.append("files", file);
        }
        formData.append("path", currentPath);

        $("#uploadProgressBar").css("width", "0%").text("0%");
        $("#uploadStatus").html("准备上传 " + this.files.length + " 个文件...");
        $("#uploadProgressModal").modal({backdrop: 'static', keyboard: false});

        let xhr = new XMLHttpRequest();
        xhr.upload.addEventListener("progress", function (e) {
            if (e.lengthComputable) {
                let percent = Math.round((e.loaded / e.total) * 100);
                $("#uploadProgressBar").css("width", percent + "%").text(percent + "%");
            }
        }, false);

        xhr.addEventListener("load", function () {
            try {
                let response = JSON.parse(xhr.responseText);
                if (response.status === "success") {
                    $("#uploadStatus").html("上传完成！");
                    setTimeout(function () {
                        $("#uploadProgressModal").modal('hide');
                        loadFileList(currentPath);
                    }, 1000);
                } else {
                    $("#uploadStatus").html("上传失败: " + response.msg);
                }
            } catch (e) {
                $("#uploadStatus").html("上传失败: 解析响应出错");
            }
        }, false);

        xhr.addEventListener("error", function () {
            $("#uploadStatus").html("上传失败: 网络错误");
        }, false);

        xhr.addEventListener("abort", function () {
            $("#uploadStatus").html("上传已取消");
        }, false);

        xhr.open("POST", "pocfile-upload", true);
        xhr.send(formData);
    });

    // 取消上传
    $("#btnCancelUpload").click(function () {
        if (xhr) xhr.abort();
        $("#uploadProgressModal").modal('hide');
    });

    // 搜索按钮
    $("#btnSearch").click(function () {
        let keyword = $("#searchKeyword").val().trim();
        let mode = $("#searchMode").val();

        if (!keyword) {
            showAlert("请输入搜索关键词", "error");
            return;
        }

        $.ajax({
            url: "pocfile-search",
            type: "GET",
            data: {
                keyword: keyword,
                mode: mode
            },
            success: function (response) {
                if (response.status === "success") {
                    renderFileList(response.data);
                    $("#currentPath").val("搜索结果");
                    currentPath = "";
                } else {
                    showAlert(response.msg, "error");
                }
            },
            error: function () {
                showAlert("搜索失败", "error");
            }
        });
    });

    // 全选/取消全选
    $("#selectAll").click(function () {
        $(".file-checkbox").prop('checked', $(this).prop('checked'));
    });

    // 删除选中文件
    $("#btnDeleteSelected").click(function () {
        let selectedFiles = [];
        $(".file-checkbox:checked").each(function () {
            selectedFiles.push($(this).data('path'));
        });

        if (selectedFiles.length === 0) {
            showAlert("请选择要删除的文件", "error");
            return;
        }


        swal({
                title: "确认删除?",
                text: "将删除选中的 " + selectedFiles.length + " 个文件/目录，此操作不可恢复！",
                type: "warning",
                showCancelButton: true,
                confirmButtonColor: "#DD6B55",
                confirmButtonText: "确认删除",
                cancelButtonText: "取消",
                closeOnConfirm: true
            },
            function () {
                $.ajax({
                    url: "pocfile-delete",
                    type: "POST",
                    data: {
                        paths: selectedFiles
                    },
                    success: function (response) {
                        if (response.status === "success") {
                            showAlert("删除成功", "success");
                            loadFileList(currentPath);
                        } else {
                            showAlert(response.msg, "error");
                        }
                    },
                    error: function () {
                        showAlert("删除失败", "error");
                    }
                });
            });
    });


// 加载文件列表
    function loadFileList(path) {
        $.ajax({
            url: "pocfile-list",
            type: "POST",
            data: {
                path: path
            },
            success: function (response) {
                if (response.status === "success") {
                    renderFileList(response.data);
                    currentPath = path;
                    $("#currentPath").val(path || "/");
                } else {
                    showAlert(response.msg, "error");
                }
            },
            error: function () {
                showAlert("加载文件列表失败", "error");
            }
        });
    }

// 渲染文件列表
    function renderFileList(files) {
        let tbody = $("#fileTableBody");
        tbody.empty();

        if (files.length === 0) {
            tbody.append('<tr><td colspan="6" class="text-center">暂无文件</td></tr>');
            return;
        }

        // 将文件和目录分开
        let directories = [];
        let fileItems = [];

        files.forEach(function(file) {
            if (file.is_dir) {
                directories.push(file);
            } else {
                fileItems.push(file);
            }
        });

        // 先渲染目录，再渲染文件
        directories.concat(fileItems).forEach(function(file) {
            let size = file.is_dir ? "-" : formatFileSize(file.size);
            let icon = file.is_dir
                ? '<i class="fa fa-folder text-warning"></i>'
                : getFileIcon(file.extension);

            let row = $('<tr>');
            row.append('<td><input type="checkbox" class="file-checkbox" data-path="' + file.path + '"></td>');

            let nameCell = $('<td>');
            let nameLink = $('<a href="#" class="file-link">').html(icon + ' ' + file.name);
            nameLink.data('file', file);
            nameCell.append(nameLink);
            row.append(nameCell);

            row.append('<td>' + size + '</td>');
            row.append('<td>' + file.mod_time + '</td>');
            row.append('<td>' + (file.md5 || '-') + '</td>');

            let actionCell = $('<td>');
            if (!file.is_dir) {
                actionCell.append('<button class="btn btn-sm btn-success btn-download" data-path="' + file.path + '" title="下载"><i class="fa fa-download"></i></button> ');
            }
            actionCell.append('<button class="btn btn-sm btn-danger btn-delete" data-path="' + file.path + '" title="删除"><i class="fa fa-trash"></i></button>');
            row.append(actionCell);

            tbody.append(row);
        });

        // 绑定事件
        $(".file-link").click(function(e) {
            e.preventDefault();
            let file = $(this).data('file');
            if (file.is_dir) {
                loadFileList(file.path);
            }
        });

        $(".btn-download").click(function() {
            let path = $(this).data('path');
            downloadFile(path);
        });

        $(".btn-delete").click(function() {
            let path = $(this).data('path');
            deleteFile(path);
        });
    }

// 下载文件
    function downloadFile(path) {
        let link = document.createElement('a');
        link.href = 'pocfile-download?path=' + encodeURIComponent(path);
        link.download = path.split('/').pop();
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    }

// 删除文件
    function deleteFile(path) {
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
                    url: "pocfile-delete",
                    type: "POST",
                    data: {
                        paths: [path]
                    },
                    success: function (response) {
                        if (response.status === "success") {
                            showAlert("删除成功", "success");
                            loadFileList(currentPath);
                        } else {
                            showAlert(response.msg, "error");
                        }
                    },
                    error: function () {
                        showAlert("删除失败", "error");
                    }
                });
            });
    }

// 获取文件图标
    function getFileIcon(ext) {
        switch (ext.toLowerCase()) {
            case '.yaml':
            case '.yml':
                return '<i class="fa fa-file-code-o text-success"></i>';
            case '.json':
                return '<i class="fa fa-file-code-o text-primary"></i>';
            case '.txt':
                return '<i class="fa fa-file-text-o text-info"></i>';
            case '.csv':
                return '<i class="fa fa-file-excel-o text-success"></i>'; // 使用Excel图标表示表格数据
            case '.js':
                return '<i class="fa fa-file-code-o text-warning"></i>'; // 黄色表示JS文件
            case '.xml':
                return '<i class="fa fa-file-code-o text-secondary"></i>'; // 灰色表示XML
            case '.svg':
                return '<i class="fa fa-file-image-o text-primary"></i>'; // 使用图片图标表示矢量图
            default:
                return '<i class="fa fa-file-o text-secondary"></i>';
        }
    }

// 格式化文件大小
    function formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

// 显示提示信息
    function showAlert(message, type) {
        swal({
            title: "Poc文件管理提示",
            text: message,
            icon: type,
            button: "确定",
        });
    }
});