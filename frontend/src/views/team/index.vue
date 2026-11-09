<template>
  <div class="team-page">
    <!-- 团队信息 -->
    <div class="card-box mb-16">
      <div class="flex-between">
        <div>
          <h3 class="page-title">{{ team.name || '我的团队' }}</h3>
          <p class="team-desc">{{ team.description || '暂无描述' }}</p>
        </div>
        <div v-if="isAdmin">
          <el-button type="primary" size="small" @click="showEditTeam = true">编辑团队</el-button>
        </div>
      </div>
    </div>

    <!-- 成员列表 -->
    <div class="card-box">
      <div class="flex-between mb-16">
        <h4>团队成员（{{ members.length }}）</h4>
        <el-button v-if="isAdmin" type="primary" size="small" @click="showInvite = true">
          <el-icon><Plus /></el-icon>&nbsp;邀请成员
        </el-button>
      </div>
      <el-table :data="members" border v-loading="loading" size="small">
        <el-table-column label="成员" min-width="180">
          <template #default="{ row }">
            <div class="member-cell">
              <el-avatar :size="32">{{ row.nickname?.charAt(0) || '?' }}</el-avatar>
              <div>
                <div class="member-name">{{ row.nickname }}</div>
                <div class="text-muted">{{ row.phone }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="USER_ROLE_MAP[row.role]?.type" size="small">
              {{ USER_ROLE_MAP[row.role]?.label || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="加入时间" width="160">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column v-if="isAdmin" label="操作" width="100">
          <template #default="{ row }">
            <el-button
              v-if="row.role !== 1"
              link
              type="danger"
              size="small"
              @click="handleRemove(row)"
            >移除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 编辑团队弹窗 -->
    <el-dialog v-model="showEditTeam" title="编辑团队" width="480px">
      <el-form label-width="80px">
        <el-form-item label="团队名称">
          <el-input v-model="editTeamForm.name" placeholder="输入团队名称" />
        </el-form-item>
        <el-form-item label="团队描述">
          <el-input v-model="editTeamForm.description" type="textarea" :rows="3" placeholder="输入团队描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditTeam = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleEditTeam">保存</el-button>
      </template>
    </el-dialog>

    <!-- 邀请成员弹窗 -->
    <el-dialog v-model="showInvite" title="邀请成员" width="480px">
      <el-form ref="inviteFormRef" :model="inviteForm" :rules="inviteRules" label-width="80px">
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="inviteForm.phone" placeholder="输入被邀请人手机号" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="inviteForm.role" style="width: 100%">
            <el-option label="管理员" :value="2" />
            <el-option label="成员" :value="3" />
            <el-option label="访客" :value="4" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showInvite = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleInvite">邀请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTeam, updateTeam, inviteMember, removeMember } from '@/api/team'
import { useUserStore } from '@/store/user'
import { USER_ROLE_MAP } from '@/constants'
import { formatDateTime } from '@/utils'

const { state: userState, isAdmin, setTeamInfo } = useUserStore()

const loading = ref(false)
const saving = ref(false)
const team = ref({})
const members = ref([])
const showEditTeam = ref(false)
const showInvite = ref(false)
const inviteFormRef = ref()

const editTeamForm = reactive({ name: '', description: '' })
const inviteForm = reactive({ phone: '', role: 3 })
const inviteRules = {
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }],
}

async function handleEditTeam() {
  saving.value = true
  try {
    await updateTeam({ name: editTeamForm.name, description: editTeamForm.description })
    ElMessage.success('更新成功')
    showEditTeam.value = false
    await loadTeam()
  } finally {
    saving.value = false
  }
}

async function handleInvite() {
  await inviteFormRef.value.validate()
  saving.value = true
  try {
    await inviteMember({ phone: inviteForm.phone, role: inviteForm.role })
    ElMessage.success('邀请成功')
    showInvite.value = false
    inviteForm.phone = ''
    inviteForm.role = 3
    await loadTeam()
  } finally {
    saving.value = false
  }
}

async function handleRemove(row) {
  await ElMessageBox.confirm(`确定移除成员「${row.nickname}」吗？`, '警告', { type: 'warning' })
  await removeMember(row.id)
  ElMessage.success('已移除')
  await loadTeam()
}

async function loadTeam() {
  loading.value = true
  try {
    const res = await getTeam()
    team.value = res.data || {}
    members.value = res.data?.members || []
    setTeamInfo(res.data)
    editTeamForm.name = team.value.name || ''
    editTeamForm.description = team.value.description || ''
  } finally {
    loading.value = false
  }
}

onMounted(loadTeam)
</script>

<style lang="scss" scoped>
.page-title {
  font-size: 20px;
  font-weight: 700;
  margin: 0;
}

.team-desc {
  font-size: 13px;
  color: #909399;
  margin: 4px 0 0;
}

.member-cell {
  display: flex;
  align-items: center;
  gap: 10px;

  .member-name {
    font-weight: 500;
    font-size: 14px;
  }
}
</style>