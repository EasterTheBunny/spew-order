<script lang="ts">
  import Paper, { Title, Content } from '@smui/paper';
  import Hero from '$lib/hero-section';
  import ProjectForm from '$lib/components/forms';
</script>

<Hero style="align-items: center; justify-content: center;">
  <Paper color="primary" elevation={12} style="width: 70%;">
    <Title>Create a new Project</Title>
    <Content>
      <div style="margin-top: 25px;">
        <ProjectForm />
      </div>
    </Content>
  </Paper>
</Hero>
