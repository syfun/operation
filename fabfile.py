# coding=utf-8

from fabric.api import lcd, settings, local, run, sudo, put, cd, env


def get_bool(value):
    if value in ['0', 0, 'False', False]:
        return False
    elif value in ['1', 1, 'True', True]:
        return True
    else:
        return False


def deploy(tmp_path, backend_url, backend_branch, ext, path, include,
           front_url, front_branch, remote_path, venv_path, program,
           workers, worker_class, bind, user_group,
           local_user, local_password, config_name='default', nginx=False):
    nginx = get_bool(nginx)

    with settings(warn_only=True):
        sudo('mkdir -p {}'.format(remote_path))
    with settings(warn_only=True):
        local('mkdir -p {}'.format(tmp_path))
    handle_backend(tmp_path, backend_url, backend_branch, remote_path,
                   user_group, venv_path, config_name=config_name)
    if front_url != 'N/A' and front_branch != 'N/A':
        handle_front(tmp_path, front_url, front_branch, remote_path, user_group,
                     local_user, local_password)
    project = '{remote_path}/backend'.format(remote_path=remote_path)
    config_supervisor(program, venv_path, project, env.user, tmp_path,
                      ext, path, include, workers, worker_class, bind)
    if nginx:
        config_nginx(remote_path, bind_host, bind_port)


supervisor_conf = """[program:{program}]
command={gunicorn} -w {workers} -k {worker_class} -b {bind} runserver:app

directory={project}
user={user}

autostart=true
autorestart=true

stdout_logfile={project}/info.log
stderr_logfile={project}/error.log
"""


def handle_backend(tmp_path, url, branch, remote_path, user_group, venv_path,
                   config_name='default'):
    """
    获取后端代码, 打包, 上传到服务器, 解压
    """

    with lcd(tmp_path):
        # 清理
        with settings(warn_only=True):
            local('rm -r backend*')

        clone_cmd = 'git clone {url} -b {branch} backend'.format(
            url=url, branch=branch)
        with settings(warn_only=True):
            res = local(clone_cmd)
        if res.failed:
            with lcd('backend'):
                local('git pull origin {branch}'.format(branch=branch))
        with lcd('backend'):
            local('rm -rf .git tools .gitignore')

        # 打包
        local('zip -r backend.zip backend')

        # 上传
        put(local_path='backend.zip',
            remote_path=remote_path,
            use_sudo=True)

        # 清理
        local('rm -r backend*')
    with cd(remote_path):
        sudo('unzip -o backend.zip')
        sudo('chown -R {} backend'.format(user_group))

        # 安装依赖环境
        with cd('backend'):
            source = 'source {venv_path}/bin/activate'.format(
                venv_path=venv_path)
            # pip_install = ('pip install -r requirements.txt '
            #                '--index-url=https://pypi.douban.com/simple')
            pip_install = 'pip install -r requirements.txt'
            cmd = '{source} && {pip_install}'.format(
                source=source, pip_install=pip_install)
            run(cmd)
            if config_name != 'default':
                run("sed -i 's/default/{}/g' runserver.py".format(config_name))


def handle_front(tmp_path, url, branch, remote_path, user_group,
                 local_user, local_password):
    """
    获取前端代码，压缩打包, 上传到服务器, 解压
    """
    with lcd(tmp_path):
        clone_cmd = 'git clone {url} -b {branch} front'.format(
            url=url, branch=branch)
        with settings(warn_only=True):
            res = local(clone_cmd)
        if res.failed:
            with lcd('front'):
                local('git checkout .')
                local('git pull origin {branch}'.format(branch=branch))

        # 修改常量
        with settings(warn_only=True):
            with lcd('front/front/src/app/core'):
                cmd = ("sed -r "
                       "-e \"s/.*fileServer.*/\.constant('fileServer', '\/api\/file\/v1')/\" "
                       "-e \"s/frontVersion/{front_branch}/2\" "
                       "-e \"s/backendVersion/{backend_branch}/2\" "
                       "constants.js > constants.js.bak").format(front_branch=front_branch, backend_branch=backend_branch)
                local(cmd)
                local('mv constants.js.bak constants.js')

        # 压缩
        with lcd('front'):
            with lcd('front'):
                # 安装gulp和bower
                with settings(warn_only=True):
                    res = local('gulp --version')
                if res.failed:
                    with settings(host_string='127.0.0.1', user=local_user,
                                  password=local_password):
                        sudo('npm install -g gulp --registry='
                             'https://registry.npm.taobao.org')
                with settings(warn_only=True):
                    res = local('bower --version')
                if res.failed:
                    with settings(host_string='127.0.0.1', user=local_user,
                                  password=local_password):
                        sudo('npm install -g bower --registry='
                             'https://registry.npm.taobao.org')

                # 安装依赖包
                local('npm install --registry=https://registry.npm.taobao.org')
                with settings(warn_only=True):
                    local('bower install')
                local('npm run build')

                with lcd('dist'):
                    # 打包压缩文件
                    local('zip -r static.zip *')
                    # 上传
                    put(local_path='static.zip',
                        remote_path=remote_path,
                        use_sudo=True)
    with cd(remote_path):
        sudo('unzip -o static.zip -d static')
        sudo('chown -R {} static'.format(user_group))


def config_supervisor(
        program, venv_path, project, user, tmp_path,
        ext='conf',
        path='/etc/supervisor/supervisord.conf',
        include='/etc/supervisor/conf.d',
        workers=4,
        worker_class='gevent',
        bind='0.0.0.0:10005'):
    gunicorn = '{venv_path}/bin/gunicorn'.format(venv_path=venv_path)
    with lcd(tmp_path):
        conf = supervisor_conf.format(
            gunicorn=gunicorn,
            project=project,
            user=user,
            program=program,
            workers=workers,
            worker_class=worker_class,
            bind=bind
        )
        file_name = '{tmp_path}/{program}.{ext}'.format(
            tmp_path=tmp_path, program=program, ext=ext)

        with open(file_name, 'w') as f:
            f.write(conf)

        put(local_path=file_name,
            remote_path=include,
            use_sudo=True)
    start_cmd = 'supervisorctl -c {conf_file} restart {program}'.format(
        conf_file=path,
        program=program
    )
    print start_cmd
    sudo(start_cmd)


def config_nginx(remote_path, host, port):
    static_path = '{}/static'.format(remote_path)
    plm_host = 'http://{host}:{port}'.format(
        host=host, port=port
    )
    file_host = 'http://119.29.75.143:10002'
    cms_host = 'http://127.0.0.1:10003'
    cmd = ("sed "
           "-e 's/$static_path/{static_path}/g' "
           "-e 's/$plm_host/{plm_host}/g' "
           "-e 's/$file_host/{file_host}/g' "
           "-e 's/$cms_host/{cms_host}/g' "
           "nginx-config > default")
    cmd = cmd.format(
        static_path=static_path.replace('/', '\/'),
        plm_host=plm_host.replace('/', '\/'),
        cms_host=cms_host.replace('/', '\/'),
        file_host=file_host.replace('/', '\/')
    )
    with cd(remote_path):
        with cd('backend'):
            run(cmd)
            sudo('mv default /etc/nginx/sites-available/')
            sudo('service nginx restart')
