import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, List, Space, Tag, Typography, message } from 'antd'

const { Title, Text } = Typography

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080'

type Todo = {
  id: number
  title: string
  completed: boolean
  created_at: string
}

function App() {
  const [form] = Form.useForm<{ title: string }>()
  const [todos, setTodos] = useState<Todo[]>([])
  const [loading, setLoading] = useState(false)
  const [api, contextHolder] = message.useMessage()

  const loadTodos = useCallback(async () => {
    setLoading(true)
    try {
      const response = await fetch(`${API_BASE_URL}/api/todos`)
      if (!response.ok) {
        throw new Error(`加载失败: ${response.status}`)
      }
      const data = (await response.json()) as Todo[]
      setTodos(data)
    } catch (error) {
      api.error((error as Error).message)
    } finally {
      setLoading(false)
    }
  }, [api])

  const addTodo = async (values: { title: string }) => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/todos`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      })
      if (!response.ok) {
        throw new Error(`创建失败: ${response.status}`)
      }
      form.resetFields()
      await loadTodos()
      api.success('Todo 创建成功')
    } catch (error) {
      api.error((error as Error).message)
    }
  }

  useEffect(() => {
    void loadTodos()
  }, [loadTodos])

  return (
    <main className="mx-auto min-h-screen w-full max-w-3xl p-6 md:p-10">
      {contextHolder}
      <Space direction="vertical" size="large" className="w-full">
        <div>
          <Title level={2} className="!mb-2">
            Fullstack Starter
          </Title>
          <Text type="secondary">
            React + Ant Design + TailwindCSS / Go + Gin + Gorm + MySQL
          </Text>
        </div>

        <Card title="创建 Todo">
          <Form form={form} layout="inline" onFinish={addTodo}>
            <Form.Item
              className="!mb-3 w-full md:!mb-0 md:!w-[70%]"
              name="title"
              rules={[
                { required: true, message: '请输入内容' },
                { min: 2, message: '至少 2 个字符' },
              ]}
            >
              <Input placeholder="输入待办事项" />
            </Form.Item>
            <Form.Item className="!mb-0">
              <Button type="primary" htmlType="submit">
                新增
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <Card
          title="Todo 列表"
          extra={
            <Button loading={loading} onClick={() => void loadTodos()}>
              刷新
            </Button>
          }
        >
          <List
            loading={loading}
            dataSource={todos}
            locale={{ emptyText: '暂无数据' }}
            renderItem={(todo) => (
              <List.Item>
                <Space>
                  <Tag color={todo.completed ? 'green' : 'blue'}>
                    {todo.completed ? '已完成' : '待处理'}
                  </Tag>
                  <span>{todo.title}</span>
                </Space>
              </List.Item>
            )}
          />
        </Card>
      </Space>
    </main>
  )
}

export default App
