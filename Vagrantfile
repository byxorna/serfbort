#Bootstrap script for installing Go and setting correct environments
BOOTSTRAP_SCRIPT = <<_SCRIPT_
GO_VERSION=1.4.2
echo 'Updating and installing Ubuntu packages...'
sudo apt-get update && sudo apt-get install -y vim git
echo 'Downloading go$GO_VERSION.linux-amd64.tar.gz'
wget --quiet -nv https://storage.googleapis.com/golang/go$GO_VERSION.linux-amd64.tar.gz
echo 'Unpacking go language'
sudo tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz
echo 'Setting up correct env. variables'
#echo "export GOPATH=/go/" >> /home/vagrant/.bashrc
echo "export GOPATH=/go/" >> /etc/profile
#echo "export PATH=$PATH:$GOPATH/bin:/usr/local/go/bin" >> /home/vagrant/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin:/usr/local/go/bin' >> /etc/profile
sudo mkdir -p /go/src/github.com/byxorna/
sudo ln -s /vagrant /go/src/github.com/byxorna/serfbort
bash -lc "cd /vagrant && sudo make setup"
_SCRIPT_

Vagrant.configure("2") do |config|
  config.vm.synced_folder ".", "/vagrant/"
  config.vm.define "master", primary: true do |m|
    m.vm.box = "ubuntu/trusty64"
    m.vm.provision :shell, inline: BOOTSTRAP_SCRIPT
  end

  #(0..10).each do |i|
  #  config.vm.define "slave#{i}" do |m|
  #    m.vm.box = "ubuntu/trusty64"
  #    m.vm.provision :shell, inline: BOOTSTRAP_SCRIPT
  #    m.vm.provision :shell, inline: "make build"
  #  end
  #end

end

