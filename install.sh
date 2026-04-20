#!/bin/sh
# MSA 一键安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/dragonTalon/msa/main/install.sh | sh
# 或在项目根目录直接运行: sh install.sh

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
info() {
    printf "${BLUE}[INFO]${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}[SUCCESS]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1"
    exit 1
}

# 检测操作系统
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "darwin" ;;
        Linux*)     echo "linux" ;;
        CYGWIN*|MINGW*|MSYS*)    echo "windows" ;;
        *)          error "不支持的操作系统: $(uname -s)" ;;
    esac
}

# 检测 CPU 架构
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)    echo "amd64" ;;
        arm64|aarch64)   echo "arm64" ;;
        *)               error "不支持的 CPU 架构: $(uname -m)" ;;
    esac
}

# 获取最新版本号（包括预发布版本）
get_latest_version() {
    # 先把响应存到变量，再处理，避免 grep -m1 关闭管道导致 curl 报错
    response=$(curl -sSL --max-time 10 https://api.github.com/repos/dragonTalon/msa/releases 2>/dev/null) || true
    version=$(echo "$response" | grep -m1 '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    echo "$version"
}

# 确定安装目录
get_install_dir() {
    # 优先使用 ~/.local/bin
    local_bin="$HOME/.local/bin"
    if [ -d "$local_bin" ] || [ -w "$(dirname "$local_bin")" ]; then
        echo "$local_bin"
        return
    fi

    # 创建 ~/.local/bin
    if mkdir -p "$local_bin" 2>/dev/null; then
        echo "$local_bin"
        return
    fi

    # 尝试 /usr/local/bin (需要 sudo)
    if [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
        return
    fi

    # 默认使用 ~/.local/bin，稍后会提示 PATH
    mkdir -p "$local_bin" 2>/dev/null || true
    echo "$local_bin"
}

# 检查 PATH
check_path() {
    install_dir="$1"
    case ":$PATH:" in
        *":$install_dir:"*)
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# 从源码构建安装
install_from_source() {
    install_dir="$1"

    info "尝试从源码构建安装..."

    # 检查是否在项目根目录
    if [ ! -f "go.mod" ] || ! grep -q 'module msa' go.mod 2>/dev/null; then
        error "未找到项目源码（go.mod），请在 msa 项目根目录运行此脚本，或确保网络可访问 GitHub"
    fi

    # 检查 go 是否安装
    if ! command -v go >/dev/null 2>&1; then
        error "未找到 go 命令，请先安装 Go: https://golang.org/dl/"
    fi

    info "正在编译..."
    if ! go build -o msa_tmp ./main.go 2>/dev/null && ! go build -o msa_tmp . 2>/dev/null; then
        error "编译失败，请检查 Go 环境"
    fi

    install_target="$install_dir/msa"
    info "正在安装到: $install_target"

    if [ -f "$install_target" ] && [ ! -w "$install_target" ]; then
        warn "需要管理员权限安装到 $install_dir"
        if command -v sudo >/dev/null 2>&1; then
            sudo mv msa_tmp "$install_target"
            sudo chmod +x "$install_target"
        else
            error "无法写入 $install_target，请手动安装"
        fi
    else
        mv msa_tmp "$install_target"
        chmod +x "$install_target"
    fi

    success "从源码安装完成!"
}

# 下载并安装
install_msa() {
    os=$(detect_os)
    arch=$(detect_arch)

    info "正在获取最新版本..."
    version=$(get_latest_version)

    install_dir=$(get_install_dir)

    # 如果无法获取版本（私有仓库/无 Release），切换到源码构建
    if [ -z "$version" ]; then
        warn "无法从 GitHub 获取版本信息（可能是私有仓库或尚无 Release）"
        install_from_source "$install_dir"

        # 检查 PATH
        if ! check_path "$install_dir"; then
            echo ""
            warn "安装目录 $install_dir 不在 PATH 中"
            echo ""
            echo "请运行以下命令添加到 PATH:"
            echo ""
            if [ -n "$(echo $SHELL | grep zsh)" ]; then
                echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
                echo "  source ~/.zshrc"
            elif [ -n "$(echo $SHELL | grep bash)" ]; then
                echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
                echo "  source ~/.bashrc"
            else
                echo "  export PATH=\"$install_dir:\$PATH\""
            fi
            echo ""
        fi

        if command -v msa >/dev/null 2>&1; then
            echo ""
            info "已安装版本:"
            msa version
        else
            echo ""
            info "请重新打开终端或运行: export PATH=\"$install_dir:\$PATH\""
            info "然后运行: msa version"
        fi
        return
    fi

    info "检测到系统: $os/$arch"
    info "最新版本: $version"

    # 构建下载 URL
    version_clean=$(echo "$version" | sed 's/^v//')
    ext="tar.gz"
    if [ "$os" = "windows" ]; then
        ext="zip"
    fi
    filename="msa_${version_clean}_${os}_${arch}.${ext}"
    download_url="https://github.com/dragonTalon/msa/releases/download/${version}/${filename}"

    info "下载地址: $download_url"

    # 创建临时目录
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # 下载文件
    tmp_file="$tmp_dir/$filename"
    info "正在下载..."
    if ! curl -fsSL --progress-bar -o "$tmp_file" "$download_url"; then
        error "下载失败"
    fi

    # 解压
    info "正在解压..."
    cd "$tmp_dir"
    if [ "$ext" = "zip" ]; then
        if ! unzip -o "$tmp_file" >/dev/null; then
            error "解压失败"
        fi
    else
        if ! tar -xzf "$tmp_file"; then
            error "解压失败"
        fi
    fi

    # 查找二进制文件
    if [ "$os" = "windows" ]; then
        binary="msa.exe"
    else
        binary="msa"
    fi

    if [ ! -f "$binary" ]; then
        error "未找到二进制文件: $binary"
    fi

    # 安装
    install_target="$install_dir/msa"
    info "正在安装到: $install_target"

    # 检查是否有写入权限
    if [ -f "$install_target" ] && [ ! -w "$install_target" ]; then
        warn "需要管理员权限安装到 $install_dir"
        if command -v sudo >/dev/null 2>&1; then
            sudo mv "$binary" "$install_target"
            sudo chmod +x "$install_target"
        else
            error "无法写入 $install_target，请手动安装"
        fi
    else
        mv "$binary" "$install_target"
        chmod +x "$install_target"
    fi

    success "安装完成!"

    # 检查 PATH
    if ! check_path "$install_dir"; then
        echo ""
        warn "安装目录 $install_dir 不在 PATH 中"
        echo ""
        echo "请运行以下命令添加到 PATH:"
        echo ""
        if [ -n "$(echo $SHELL | grep zsh)" ]; then
            echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
            echo "  source ~/.zshrc"
        elif [ -n "$(echo $SHELL | grep bash)" ]; then
            echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
            echo "  source ~/.bashrc"
        else
            echo "  export PATH=\"$install_dir:\$PATH\""
        fi
        echo ""
    fi

    # 验证安装
    if command -v msa >/dev/null 2>&1; then
        echo ""
        info "已安装版本:"
        msa version
    else
        echo ""
        info "请重新打开终端或运行: export PATH=\"$install_dir:\$PATH\""
        info "然后运行: msa version"
    fi
}

# 主函数
main() {
    echo ""
    echo "==================================="
    echo "     MSA 一键安装脚本"
    echo "==================================="
    echo ""

    install_msa
}

main "$@"
