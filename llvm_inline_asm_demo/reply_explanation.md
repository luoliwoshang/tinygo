这个 panic 逻辑与目前 TinyGo 的行为完全一致，并且追溯到最早提交这行代码的 commit 记录 (https://github.com/tinygo-org/tinygo/commit/
  392bba8394ece9e8b90ee2d6e82590b1492bc62d)，并没有指出这里的理由。但是我认为根本的技术原因是 **LLVM
  内联汇编要求所有调用信息都必须在编译时确定** - 参数类型、约束字符串和参数数量不能在运行时动态变化。

  ```llvm
  call void asm "mov $0, $1", "r,r" (i32 %arg1, i32 %arg2)

  这本质上与任何函数调用需要在编译时确定函数签名和参数类型的要求相同。

  如果允许跨基本块的 map 构造，编译器无法确定调用需要什么参数：

  regs := make(map[string]interface{})
  if condition {
      regs["value"] = uint32(42)    // 类型1: uint32
  }
  res1 := asmFull("mov {}, {value}", regs)  // 编译器无法确定 value 是否存在或其类型

  这个约束源于 LLVM 的技术限制，而非设计偏好。