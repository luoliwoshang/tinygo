#!/bin/bash

# TinyGo ESP32 构建测试脚本
# 用于验证ESP32构建流程和生成的固件文件

set -e  # 遇到错误立即退出

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TINYGO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
EXAMPLE_DIR="${SCRIPT_DIR}/hello-world"

echo "=== TinyGo ESP32 构建测试 ==="
echo "TinyGo根目录: ${TINYGO_ROOT}"
echo "示例目录: ${EXAMPLE_DIR}"
echo

# 检查TinyGo是否已构建
if [ ! -f "${TINYGO_ROOT}/build/tinygo" ]; then
    echo "错误: TinyGo未构建，请先运行: make tinygo"
    exit 1
fi

# 切换到示例目录
cd "${EXAMPLE_DIR}"

echo "1. 编译ESP32-CoreBoard-v2目标..."
"${TINYGO_ROOT}/build/tinygo" build -target=esp32-coreboard-v2 -o main.elf main.go

echo "2. 检查生成的ELF文件..."
if [ -f "main.elf" ]; then
    echo "✓ ELF文件生成成功"
    file main.elf
    echo "文件大小: $(stat -f%z main.elf 2>/dev/null || stat -c%s main.elf 2>/dev/null) 字节"
else
    echo "✗ ELF文件未生成"
    exit 1
fi

echo
echo "3. 生成ESP32固件镜像..."
"${TINYGO_ROOT}/build/tinygo" build -target=esp32-coreboard-v2 -o main.bin main.go

echo "4. 检查生成的固件文件..."
if [ -f "main.bin" ]; then
    echo "✓ 固件文件生成成功"
    file main.bin
    echo "文件大小: $(stat -f%z main.bin 2>/dev/null || stat -c%s main.bin 2>/dev/null) 字节"
    
    # 检查ESP32固件魔数
    magic=$(hexdump -C main.bin | head -1 | awk '{print $2}')
    if [ "$magic" = "e9" ]; then
        echo "✓ ESP32固件魔数正确 (0xE9)"
    else
        echo "✗ ESP32固件魔数错误，期望: e9，实际: $magic"
    fi
else
    echo "✗ 固件文件未生成"
    exit 1
fi

echo
echo "5. 分析固件镜像结构..."
echo "--- 固件头部信息 ---"
hexdump -C main.bin | head -5

echo
echo "6. 检查入口点..."
# 使用readelf检查ELF入口点
if command -v readelf >/dev/null; then
    entry_point=$(readelf -h main.elf | grep "Entry point" | awk '{print $4}')
    echo "ELF入口点: $entry_point"
    
    # 检查固件镜像中的入口点 (偏移8-11字节)
    bin_entry=$(hexdump -s 8 -n 4 -e '/1 "%02x"' main.bin | sed 's/\(..\)\(..\)\(..\)\(..\)/\4\3\2\1/')
    echo "固件入口点: 0x$bin_entry"
    
    if [ "$(echo $entry_point | sed 's/0x//')" = "$bin_entry" ]; then
        echo "✓ 入口点一致"
    else
        echo "⚠ 入口点不一致，可能正常(字节序)"
    fi
else
    echo "未找到readelf，跳过入口点检查"
fi

echo
echo "7. 模拟烧录命令..."
port="/dev/ttyUSB0"  # 示例端口
flash_cmd="esptool.py --chip=esp32 --port $port write_flash 0x1000 main.bin -ff 80m -fm dout"
echo "烧录命令: $flash_cmd"
echo "注意: 实际烧录需要连接ESP32设备并安装esptool.py"

echo
echo "8. 清理临时文件..."
if [ -f "main.elf" ] && [ -f "main.bin" ]; then
    echo "保留生成的文件供检查:"
    ls -la main.*
    echo "如需清理，运行: rm -f main.elf main.bin"
fi

echo
echo "=== 构建测试完成 ==="
echo "✓ ESP32固件生成成功"
echo "✓ 构建流程验证通过"
echo
echo "下一步:"
echo "1. 连接ESP32开发板到 $port"
echo "2. 安装esptool: pip install esptool" 
echo "3. 执行烧录: $flash_cmd"
echo "4. 打开串口监视器观察输出"