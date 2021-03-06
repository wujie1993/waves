# 参与项目

## 问题处理流程

```mermaid
sequenceDiagram
    报告者-->>问题: 1. 创建问题
    Note right of 问题: ① 运行环境<br/>② 问题复现<br/>③ 预期结果<br/>④ 标签phase: todo
    提交者-->>问题: 2. 指定问题处理并更新标签
    Note right of 问题: 标签phase: doing
    提交者-->>仓库: 3. 发起提交或PR
    Note right of 提交者: 提交附带问题编号
    提交者-->>问题: 4. 更新标签
    Note right of 问题: 标签phase: done
    提交者-->>审阅者: 5. 指定问题审阅
    Note right of 审阅者: 审阅并验证结果<br/>不通过则返回步骤2
    审阅者-->>问题: 6. 更新标签
    Note right of 问题: 标签lgtm
    审阅者-->>维护者: 7. 指定问题审阅
    Note right of 维护者: 审阅不通过<br/>则返回步骤2
    维护者-->>问题: 8. 关闭issue
    维护者-->>提交者: 9. 指定问题最终完成者
```

**备注**

- 问题报告者(reporter)可以是任何成员
- 提交者提交完代码后需要在issue中说明验证的方法
- 审阅者一般选择除提交者外较为熟悉相关代码模块的成员
- 审阅者需要根据提交者提供的验证方法验证结果

## 问题模板

**bug模板**

```
**发生了什么**:
<!-- 此处补充问题产生的结果 -->

**预期的结果**:
<!-- 此处补充正常情况下应该是什么结果 -->

**如何重现**:
<!-- 此处补充复现问题的步骤 -->

**环境**:

- 操作系统: 
- 程序版本:
- 机器配置:
```

**新特性模板**

```
**哪些方面的提升**
<!-- 此处补充该新特性所带来的好处 -->

**原因或需求**
<!-- 此处补充实现该新特性的缘由或现实需求 -->
```

## 提交信息格式

```
{{ 所属模块 }} {{ issue编号 }}: {{ 简单描述 }}

{{ 详细描述 }}

Signed-off-by: {{ 用户名 }} <{{ 用户邮箱 }}>
```
